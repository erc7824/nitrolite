package stress

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Run is the main entry point for the stress-test subcommand.
// It receives os.Args[2:] (everything after "stress-test").
// Returns exit code: 0 if all pass, 1 if any fail.
func Run(args []string) int {
	if len(args) == 0 {
		printUsage()
		return 1
	}

	cfg, err := ReadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}

	walletAddress, err := cfg.WalletAddress()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}
	if os.Getenv("STRESS_PRIVATE_KEY") == "" {
		fmt.Printf("Wallet: %s (ephemeral)\n", walletAddress)
	} else {
		fmt.Printf("Wallet: %s\n", walletAddress)
	}

	specs, err := parseSpecs(args, cfg.Connections)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}

	maxConns := 0
	for _, s := range specs {
		if s.Connections > maxConns {
			maxConns = s.Connections
		}
	}

	fmt.Printf("Opening %d WebSocket connections to %s...\n", maxConns, cfg.WsURL)
	clients, err := CreateClientPool(cfg.WsURL, cfg.PrivateKey, maxConns)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to create connection pool: %v\n", err)
		return 1
	}
	defer CloseClientPool(clients)

	registry := MethodRegistry()
	allPassed := true

	for i, spec := range specs {
		fmt.Printf("\n[%d/%d] Running: %s (%d requests, %d connections)\n",
			i+1, len(specs), spec.Method, spec.TotalReqs, spec.Connections)

		factory, ok := registry[spec.Method]
		if !ok {
			fmt.Fprintf(os.Stderr, "ERROR: Unknown method: %s\nAvailable: %s\n",
				spec.Method, strings.Join(sortedMethodNames(), ", "))
			allPassed = false
			continue
		}

		fn, err := factory(spec.ExtraArgs, walletAddress)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			allPassed = false
			continue
		}

		poolSlice := clients[:spec.Connections]

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
		results, totalTime := RunTest(ctx, spec.TotalReqs, poolSlice, fn)
		cancel()

		report := ComputeReport(spec.Method, spec.TotalReqs, spec.Connections, results, totalTime)
		PrintReport(report)

		errorRate := float64(report.Failed) / float64(report.TotalReqs)
		if errorRate > cfg.MaxErrorRate {
			fmt.Printf("FAIL: Error rate %.2f%% exceeds threshold %.2f%%\n",
				errorRate*100, cfg.MaxErrorRate*100)
			allPassed = false
		} else {
			fmt.Printf("PASS: Error rate %.2f%% within threshold %.2f%%\n",
				errorRate*100, cfg.MaxErrorRate*100)
		}
	}

	if allPassed {
		fmt.Println("\nAll stress tests PASSED.")
		return 0
	}
	fmt.Println("\nSome stress tests FAILED.")
	return 1
}

// parseSpecs parses CLI arguments in format "method:total_requests[:connections[:extra_params...]]"
func parseSpecs(args []string, defaultConns int) ([]TestSpec, error) {
	specs := make([]TestSpec, 0, len(args))
	for _, arg := range args {
		parts := strings.Split(arg, ":")
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid spec %q: expected method:total_requests[:connections[:params...]]", arg)
		}

		method := parts[0]
		totalReqs, err := strconv.Atoi(parts[1])
		if err != nil || totalReqs <= 0 {
			return nil, fmt.Errorf("invalid total_requests in %q: must be positive integer", arg)
		}

		connections := defaultConns
		extraStart := 2
		if len(parts) >= 3 {
			c, err := strconv.Atoi(parts[2])
			if err == nil && c > 0 {
				connections = c
				extraStart = 3
			}
			// If not a valid int, treat it as an extra param (e.g. wallet address)
		}

		var extra []string
		if len(parts) > extraStart {
			extra = parts[extraStart:]
		}

		specs = append(specs, TestSpec{
			Method:      method,
			TotalReqs:   totalReqs,
			Connections: connections,
			ExtraArgs:   extra,
		})
	}
	return specs, nil
}

func printUsage() {
	fmt.Println("Usage: clearnode stress-test <spec> [<spec>...]")
	fmt.Println()
	fmt.Println("Spec format: method:total_requests[:connections[:extra_params...]]")
	fmt.Println()
	fmt.Println("Environment variables:")
	fmt.Println("  STRESS_WS_URL          (required) WebSocket URL of the clearnode")
	fmt.Println("  STRESS_PRIVATE_KEY     (optional) Hex private key for signing (ephemeral if not set)")
	fmt.Println("  STRESS_CONNECTIONS     (optional) Default connections per test (default: 10)")
	fmt.Println("  STRESS_TIMEOUT         (optional) Per-test timeout (default: 10m)")
	fmt.Println("  STRESS_MAX_ERROR_RATE  (optional) Max error rate threshold (default: 0.01)")
	fmt.Println()
	fmt.Println("Methods:", strings.Join(sortedMethodNames(), ", "))
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  clearnode stress-test ping:1000:10")
	fmt.Println("  clearnode stress-test ping:1000:10 get-balances:2000:20:0xWALLET")
	fmt.Println("  clearnode stress-test get-config:500")
}

func sortedMethodNames() []string {
	registry := MethodRegistry()
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
