package main

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/sign"
	sdk "github.com/erc7824/nitrolite/sdk/go"
)

// ============================================================================
// Stress Test Types
// ============================================================================

// stressMethodFunc is the signature for a single stress test invocation.
type stressMethodFunc func(ctx context.Context, client *sdk.Client) error

// stressFactory parses method-specific args and returns a stressMethodFunc.
type stressFactory func(args []string) (stressMethodFunc, error)

// StressResult captures the outcome of a single request.
type StressResult struct {
	Duration time.Duration
	Err      error
}

// StressReport contains the full aggregated report after a stress test run.
type StressReport struct {
	Method      string
	TotalReqs   int
	Connections int
	Successful  int
	Failed      int
	TotalTime   time.Duration

	MinLatency    time.Duration
	MaxLatency    time.Duration
	AvgLatency    time.Duration
	MedianLatency time.Duration
	P95Latency    time.Duration
	P99Latency    time.Duration

	RequestsPerSec float64
	ErrorBreakdown map[string]int
}

// ============================================================================
// Stress Job Manager
// ============================================================================

// stressJob tracks a running or completed stress test.
type stressJob struct {
	ID          int
	Method      string
	TotalReqs   int
	NumConns    int
	StartedAt   time.Time
	FinishedAt  time.Time
	Done        bool
	Report      *StressReport
	Cancel      context.CancelFunc
	Clients     []*sdk.Client

	// live counters (read atomically)
	completed int64
	failed    int64
	totalNs   int64
}

func (j *stressJob) liveRPS() float64 {
	c := atomic.LoadInt64(&j.completed)
	if c == 0 {
		return 0
	}
	elapsed := time.Since(j.StartedAt).Seconds()
	if elapsed == 0 {
		return 0
	}
	return float64(c) / elapsed
}

func (j *stressJob) liveAvg() time.Duration {
	c := atomic.LoadInt64(&j.completed)
	if c == 0 {
		return 0
	}
	return time.Duration(atomic.LoadInt64(&j.totalNs) / c)
}

// stressManager holds all stress test jobs.
type stressManager struct {
	mu     sync.Mutex
	jobs   []*stressJob
	nextID int
}

func newStressManager() *stressManager {
	return &stressManager{nextID: 1}
}

func (m *stressManager) add(job *stressJob) {
	m.mu.Lock()
	defer m.mu.Unlock()
	job.ID = m.nextID
	m.nextID++
	m.jobs = append(m.jobs, job)
}

func (m *stressManager) get(id int) *stressJob {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, j := range m.jobs {
		if j.ID == id {
			return j
		}
	}
	return nil
}

func (m *stressManager) all() []*stressJob {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]*stressJob, len(m.jobs))
	copy(out, m.jobs)
	return out
}

// ============================================================================
// Connection Pool
// ============================================================================

func (o *Operator) createClientPool(n int) ([]*sdk.Client, error) {
	privateKey, err := o.store.GetPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("no wallet imported: %w", err)
	}

	stateSigner, err := o.buildStateSigner(privateKey)
	if err != nil {
		return nil, err
	}

	txSigner, err := sign.NewEthereumRawSigner(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create tx signer: %w", err)
	}

	rpcs, err := o.store.GetAllRPCs()
	if err != nil {
		rpcs = make(map[uint64]string)
	}

	opts := []sdk.Option{
		sdk.WithErrorHandler(func(_ error) {}),
	}
	for chainID, rpcURL := range rpcs {
		opts = append(opts, sdk.WithBlockchainRPC(chainID, rpcURL))
	}

	clients := make([]*sdk.Client, 0, n)
	for i := 0; i < n; i++ {
		client, err := sdk.NewClient(o.wsURL, stateSigner, txSigner, opts...)
		if err != nil {
			for _, c := range clients {
				c.Close()
			}
			return nil, fmt.Errorf("failed to open connection %d/%d: %w", i+1, n, err)
		}
		clients = append(clients, client)
	}

	return clients, nil
}

func closeClientPool(clients []*sdk.Client) {
	for _, c := range clients {
		c.Close()
	}
}

// ============================================================================
// Generic Stress Test Runner
// ============================================================================

// runStressTest executes totalReqs calls of fn distributed across the client pool.
// It updates the job's atomic counters for live monitoring.
func runStressTest(ctx context.Context, totalReqs int, clients []*sdk.Client, fn stressMethodFunc, job *stressJob) ([]StressResult, time.Duration) {
	numClients := len(clients)
	results := make([]StressResult, totalReqs)
	sem := make(chan struct{}, totalReqs)
	if numClients < totalReqs {
		sem = make(chan struct{}, numClients*10)
	}
	var wg sync.WaitGroup

	start := time.Now()

	for i := 0; i < totalReqs; i++ {
		sem <- struct{}{}
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			defer func() { <-sem }()

			client := clients[idx%numClients]

			reqStart := time.Now()
			err := fn(ctx, client)
			d := time.Since(reqStart)
			results[idx] = StressResult{
				Duration: d,
				Err:      err,
			}

			atomic.AddInt64(&job.totalNs, int64(d))
			if err != nil {
				atomic.AddInt64(&job.failed, 1)
			}
			atomic.AddInt64(&job.completed, 1)
		}(i)
	}

	wg.Wait()
	totalTime := time.Since(start)

	return results, totalTime
}

// ============================================================================
// Report Computation
// ============================================================================

func computeReport(method string, totalReqs, connections int, results []StressResult, totalTime time.Duration) StressReport {
	report := StressReport{
		Method:         method,
		TotalReqs:      totalReqs,
		Connections:    connections,
		TotalTime:      totalTime,
		ErrorBreakdown: make(map[string]int),
	}

	durations := make([]time.Duration, 0, len(results))

	for _, r := range results {
		if r.Err != nil {
			report.Failed++
			report.ErrorBreakdown[r.Err.Error()]++
		} else {
			report.Successful++
		}
		durations = append(durations, r.Duration)
	}

	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })

	if len(durations) > 0 {
		report.MinLatency = durations[0]
		report.MaxLatency = durations[len(durations)-1]

		var total time.Duration
		for _, d := range durations {
			total += d
		}
		report.AvgLatency = total / time.Duration(len(durations))
		report.MedianLatency = durationPercentile(durations, 50)
		report.P95Latency = durationPercentile(durations, 95)
		report.P99Latency = durationPercentile(durations, 99)
	}

	if totalTime.Seconds() > 0 {
		report.RequestsPerSec = float64(totalReqs) / totalTime.Seconds()
	}

	return report
}

func durationPercentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p / 100.0)
	return sorted[idx]
}

// ============================================================================
// Report Printing
// ============================================================================

func printStressReport(report StressReport) {
	fmt.Println()
	fmt.Println("Stress Test Report")
	fmt.Println("==================")
	fmt.Printf("Method:          %s\n", report.Method)
	fmt.Printf("Total Requests:  %d\n", report.TotalReqs)
	fmt.Printf("Connections:     %d\n", report.Connections)
	fmt.Printf("Duration:        %s\n", report.TotalTime.Round(time.Millisecond))

	fmt.Println()
	fmt.Println("Results")
	fmt.Println("-------")
	successPct := float64(report.Successful) / float64(report.TotalReqs) * 100
	failPct := float64(report.Failed) / float64(report.TotalReqs) * 100
	fmt.Printf("Successful:      %d (%.1f%%)\n", report.Successful, successPct)
	fmt.Printf("Failed:          %d (%.1f%%)\n", report.Failed, failPct)
	fmt.Printf("Requests/sec:    %.2f\n", report.RequestsPerSec)

	fmt.Println()
	fmt.Println("Latency")
	fmt.Println("-------")
	fmt.Printf("Min:             %s\n", report.MinLatency.Round(time.Microsecond))
	fmt.Printf("Max:             %s\n", report.MaxLatency.Round(time.Microsecond))
	fmt.Printf("Average:         %s\n", report.AvgLatency.Round(time.Microsecond))
	fmt.Printf("Median (p50):    %s\n", report.MedianLatency.Round(time.Microsecond))
	fmt.Printf("P95:             %s\n", report.P95Latency.Round(time.Microsecond))
	fmt.Printf("P99:             %s\n", report.P99Latency.Round(time.Microsecond))

	if len(report.ErrorBreakdown) > 0 {
		fmt.Println()
		fmt.Println("Errors")
		fmt.Println("------")
		for errMsg, count := range report.ErrorBreakdown {
			fmt.Printf("  %-60s %d\n", errMsg, count)
		}
	}

	fmt.Println()
}

// ============================================================================
// Method Registry
// ============================================================================

func (o *Operator) getStressMethodRegistry() map[string]stressFactory {
	return map[string]stressFactory{
		"ping":                   o.stressPing,
		"get-config":             o.stressGetConfig,
		"get-blockchains":        o.stressGetBlockchains,
		"get-assets":             o.stressGetAssets,
		"get-balances":           o.stressGetBalances,
		"get-transactions":       o.stressGetTransactions,
		"get-home-channel":       o.stressGetHomeChannel,
		"get-escrow-channel":     o.stressGetEscrowChannel,
		"get-latest-state":       o.stressGetLatestState,
		"get-channel-key-states": o.stressGetLastChannelKeyStates,
		"get-app-sessions":       o.stressGetAppSessions,
		"get-app-key-states":     o.stressGetLastAppKeyStates,
	}
}

// ============================================================================
// Entry Points
// ============================================================================

func (o *Operator) runStressCommand(args []string) {
	if len(args) < 4 {
		fmt.Println("ERROR: Usage: stress <method> <total_requests> <connections> [method-params...]")
		fmt.Println()
		fmt.Println("  Runs in the background. Use 'stress-status' to monitor, 'stress-results <id>' to see report.")
		fmt.Println()
		fmt.Println("Methods: ping, get-config, get-blockchains, get-assets, get-balances,")
		fmt.Println("         get-transactions, get-home-channel, get-escrow-channel,")
		fmt.Println("         get-latest-state, get-channel-key-states, get-app-sessions,")
		fmt.Println("         get-app-key-states")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  stress ping 1000 10")
		fmt.Println("  stress get-balances 5000 50")
		fmt.Println("  stress get-home-channel 2000 20 usdc")
		return
	}

	methodName := args[1]

	totalReqs, err := strconv.Atoi(args[2])
	if err != nil || totalReqs <= 0 {
		fmt.Println("ERROR: total_requests must be a positive integer")
		return
	}

	numConns, err := strconv.Atoi(args[3])
	if err != nil || numConns <= 0 {
		fmt.Println("ERROR: connections must be a positive integer")
		return
	}

	registry := o.getStressMethodRegistry()
	factory, ok := registry[methodName]
	if !ok {
		fmt.Printf("ERROR: Unknown stress method: %s\n", methodName)
		fmt.Println("Available methods: " + strings.Join(o.stressMethodNames(), ", "))
		return
	}

	fn, err := factory(args[4:])
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	fmt.Printf("Opening %d WebSocket connections...\n", numConns)
	clients, err := o.createClientPool(numConns)
	if err != nil {
		fmt.Printf("ERROR: Failed to create connection pool: %v\n", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)

	job := &stressJob{
		Method:    methodName,
		TotalReqs: totalReqs,
		NumConns:  numConns,
		StartedAt: time.Now(),
		Cancel:    cancel,
		Clients:   clients,
	}
	o.stressJobs.add(job)

	fmt.Printf("Started stress test #%d: %s (%d requests, %d connections)\n", job.ID, methodName, totalReqs, numConns)
	fmt.Println("Use 'stress-status' to monitor, 'stress-results " + strconv.Itoa(job.ID) + "' for full report when done.")

	// Run in background
	go func() {
		defer cancel()
		defer closeClientPool(clients)

		results, totalTime := runStressTest(ctx, totalReqs, clients, fn, job)
		report := computeReport(methodName, totalReqs, numConns, results, totalTime)
		job.Report = &report
		job.FinishedAt = time.Now()
		job.Done = true

		fmt.Printf("\nStress test #%d (%s) completed: %.1f rps, avg=%s, errors=%d\n",
			job.ID, methodName, report.RequestsPerSec,
			report.AvgLatency.Round(time.Microsecond), report.Failed)
	}()
}

func (o *Operator) showStressStatus() {
	jobs := o.stressJobs.all()
	if len(jobs) == 0 {
		fmt.Println("No stress tests running or completed.")
		return
	}

	fmt.Println()
	fmt.Printf("  %-4s %-22s %-10s %-8s %-12s %-10s %-8s %s\n",
		"ID", "Method", "Requests", "Conns", "Progress", "RPS", "Errors", "Status")
	fmt.Printf("  %-4s %-22s %-10s %-8s %-12s %-10s %-8s %s\n",
		"--", "------", "--------", "-----", "--------", "---", "------", "------")

	for _, j := range jobs {
		completed := atomic.LoadInt64(&j.completed)
		failed := atomic.LoadInt64(&j.failed)
		pct := float64(completed) / float64(j.TotalReqs) * 100

		status := "running"
		rpsStr := fmt.Sprintf("%.1f", j.liveRPS())
		if j.Done {
			status = "done"
			if j.Report != nil {
				rpsStr = fmt.Sprintf("%.1f", j.Report.RequestsPerSec)
			}
		}

		fmt.Printf("  %-4d %-22s %-10d %-8d %5d/%d %3.0f%% %-10s %-8d %s\n",
			j.ID, j.Method, j.TotalReqs, j.NumConns,
			completed, j.TotalReqs, pct, rpsStr, failed, status)
	}
	fmt.Println()
}

func (o *Operator) showStressResults(args []string) {
	if len(args) < 2 {
		// No ID given — show all completed reports
		jobs := o.stressJobs.all()
		found := false
		for _, j := range jobs {
			if j.Done && j.Report != nil {
				found = true
				fmt.Printf("=== Stress Test #%d ===", j.ID)
				printStressReport(*j.Report)
			}
		}
		if !found {
			fmt.Println("No completed stress tests. Use 'stress-status' to check running tests.")
		}
		return
	}

	id, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("ERROR: Usage: stress-results [id]")
		return
	}

	job := o.stressJobs.get(id)
	if job == nil {
		fmt.Printf("ERROR: Stress test #%d not found\n", id)
		return
	}

	if !job.Done {
		completed := atomic.LoadInt64(&job.completed)
		failed := atomic.LoadInt64(&job.failed)
		fmt.Printf("Stress test #%d is still running: %d/%d completed (%.0f%%), avg=%s, rps=%.1f, errors=%d\n",
			id, completed, job.TotalReqs,
			float64(completed)/float64(job.TotalReqs)*100,
			job.liveAvg().Round(time.Microsecond), job.liveRPS(), failed)
		return
	}

	if job.Report == nil {
		fmt.Printf("Stress test #%d completed but no report available\n", id)
		return
	}

	fmt.Printf("=== Stress Test #%d ===", job.ID)
	printStressReport(*job.Report)
}

func (o *Operator) stopStressTest(args []string) {
	if len(args) < 2 {
		fmt.Println("ERROR: Usage: stress-stop <id>")
		return
	}

	id, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("ERROR: Usage: stress-stop <id>")
		return
	}

	job := o.stressJobs.get(id)
	if job == nil {
		fmt.Printf("ERROR: Stress test #%d not found\n", id)
		return
	}

	if job.Done {
		fmt.Printf("Stress test #%d already completed\n", id)
		return
	}

	job.Cancel()
	fmt.Printf("Stopping stress test #%d...\n", id)
}

func (o *Operator) stressMethodNames() []string {
	registry := o.getStressMethodRegistry()
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ============================================================================
// Method Factories
// ============================================================================

func (o *Operator) stressPing(_ []string) (stressMethodFunc, error) {
	return func(ctx context.Context, client *sdk.Client) error {
		return client.Ping(ctx)
	}, nil
}

func (o *Operator) stressGetConfig(_ []string) (stressMethodFunc, error) {
	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetConfig(ctx)
		return err
	}, nil
}

func (o *Operator) stressGetBlockchains(_ []string) (stressMethodFunc, error) {
	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetBlockchains(ctx)
		return err
	}, nil
}

func (o *Operator) stressGetAssets(args []string) (stressMethodFunc, error) {
	var chainID *uint64
	if len(args) >= 1 {
		parsed, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid chain_id: %s", args[0])
		}
		chainID = &parsed
	}

	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetAssets(ctx, chainID)
		return err
	}, nil
}

func (o *Operator) stressGetBalances(args []string) (stressMethodFunc, error) {
	wallet := ""
	if len(args) >= 1 {
		wallet = args[0]
	} else {
		wallet = o.getImportedWalletAddress()
		if wallet == "" {
			return nil, fmt.Errorf("no wallet configured; provide wallet address or use 'import wallet' first")
		}
	}

	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetBalances(ctx, wallet)
		return err
	}, nil
}

func (o *Operator) stressGetTransactions(args []string) (stressMethodFunc, error) {
	wallet := ""
	if len(args) >= 1 {
		wallet = args[0]
	} else {
		wallet = o.getImportedWalletAddress()
		if wallet == "" {
			return nil, fmt.Errorf("no wallet configured; provide wallet address or use 'import wallet' first")
		}
	}

	limit := uint32(20)
	opts := &sdk.GetTransactionsOptions{
		Pagination: &core.PaginationParams{
			Limit: &limit,
		},
	}

	return func(ctx context.Context, client *sdk.Client) error {
		_, _, err := client.GetTransactions(ctx, wallet, opts)
		return err
	}, nil
}

func (o *Operator) stressGetHomeChannel(args []string) (stressMethodFunc, error) {
	var wallet, asset string

	switch len(args) {
	case 2:
		wallet = args[0]
		asset = args[1]
	case 1:
		wallet = o.getImportedWalletAddress()
		if wallet == "" {
			return nil, fmt.Errorf("no wallet configured; provide wallet address or use 'import wallet' first")
		}
		asset = args[0]
	default:
		return nil, fmt.Errorf("usage: stress get-home-channel <total> <connections> [wallet] <asset>")
	}

	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetHomeChannel(ctx, wallet, asset)
		return err
	}, nil
}

func (o *Operator) stressGetEscrowChannel(args []string) (stressMethodFunc, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("usage: stress get-escrow-channel <total> <connections> <channel_id>")
	}
	channelID := args[0]

	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetEscrowChannel(ctx, channelID)
		return err
	}, nil
}

func (o *Operator) stressGetLatestState(args []string) (stressMethodFunc, error) {
	var wallet, asset string

	switch len(args) {
	case 2:
		wallet = args[0]
		asset = args[1]
	case 1:
		wallet = o.getImportedWalletAddress()
		if wallet == "" {
			return nil, fmt.Errorf("no wallet configured; provide wallet address or use 'import wallet' first")
		}
		asset = args[0]
	default:
		return nil, fmt.Errorf("usage: stress get-latest-state <total> <connections> [wallet] <asset>")
	}

	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetLatestState(ctx, wallet, asset, false)
		return err
	}, nil
}

func (o *Operator) stressGetLastChannelKeyStates(args []string) (stressMethodFunc, error) {
	wallet := ""
	if len(args) >= 1 {
		wallet = args[0]
	} else {
		wallet = o.getImportedWalletAddress()
		if wallet == "" {
			return nil, fmt.Errorf("no wallet configured; provide wallet address or use 'import wallet' first")
		}
	}

	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetLastChannelKeyStates(ctx, wallet, nil)
		return err
	}, nil
}

func (o *Operator) stressGetAppSessions(args []string) (stressMethodFunc, error) {
	wallet := ""
	if len(args) >= 1 {
		wallet = args[0]
	} else {
		wallet = o.getImportedWalletAddress()
	}

	limit := uint32(20)
	opts := &sdk.GetAppSessionsOptions{
		Pagination: &core.PaginationParams{
			Limit: &limit,
		},
	}
	if wallet != "" {
		opts.Participant = &wallet
	}

	return func(ctx context.Context, client *sdk.Client) error {
		_, _, err := client.GetAppSessions(ctx, opts)
		return err
	}, nil
}

func (o *Operator) stressGetLastAppKeyStates(args []string) (stressMethodFunc, error) {
	wallet := ""
	if len(args) >= 1 {
		wallet = args[0]
	} else {
		wallet = o.getImportedWalletAddress()
		if wallet == "" {
			return nil, fmt.Errorf("no wallet configured; provide wallet address or use 'import wallet' first")
		}
	}

	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetLastAppKeyStates(ctx, wallet, nil)
		return err
	}, nil
}
