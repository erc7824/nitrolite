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
// Returns exit code: 0 if pass, 1 if fail.
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

	spec, err := parseSpec(args[0], cfg.Connections)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}

	registry := MethodRegistry()
	runner, ok := registry[spec.Method]
	if !ok {
		fmt.Fprintf(os.Stderr, "ERROR: Unknown method: %s\nAvailable: %s\n",
			spec.Method, strings.Join(sortedMethodNames(), ", "))
		return 1
	}

	fmt.Printf("Running: %s (%d requests)\n", spec.Method, spec.TotalReqs)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	report, err := runner(ctx, cfg, spec)
	cancel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}

	PrintReport(report)

	errorRate := float64(report.Failed) / float64(report.TotalReqs)
	if errorRate > cfg.MaxErrorRate {
		fmt.Printf("FAIL: Error rate %.2f%% exceeds threshold %.2f%%\n",
			errorRate*100, cfg.MaxErrorRate*100)
		return 1
	}
	fmt.Printf("PASS: Error rate %.2f%% within threshold %.2f%%\n",
		errorRate*100, cfg.MaxErrorRate*100)
	return 0
}

// parseSpec parses a single CLI argument in format "method:total_requests[:connections[:extra_params...]]"
func parseSpec(arg string, defaultConns int) (TestSpec, error) {
	parts := strings.Split(arg, ":")
	if len(parts) < 2 {
		return TestSpec{}, fmt.Errorf("invalid spec %q: expected method:total_requests[:connections[:params...]]", arg)
	}

	method := parts[0]
	totalReqs, err := strconv.Atoi(parts[1])
	if err != nil || totalReqs <= 0 {
		return TestSpec{}, fmt.Errorf("invalid total_requests in %q: must be positive integer", arg)
	}

	connections := defaultConns
	extraStart := 2
	if len(parts) >= 3 {
		c, err := strconv.Atoi(parts[2])
		if err == nil && c > 0 {
			connections = c
			extraStart = 3
		}
		// If not a valid int, treat it as an extra param (e.g. asset name)
	}

	var extra []string
	if len(parts) > extraStart {
		extra = parts[extraStart:]
	}

	return TestSpec{
		Method:      method,
		TotalReqs:   totalReqs,
		Connections: connections,
		ExtraArgs:   extra,
	}, nil
}

func printUsage() {
	fmt.Println("Usage: clearnode stress-test <spec>")
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
	fmt.Println("Available methods:", strings.Join(sortedMethodNames(), ", "))
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  clearnode stress-test ping:1000:10")
	fmt.Println("  clearnode stress-test get-balances:2000:20:0xWALLET")
	fmt.Println("  clearnode stress-test transfer-roundtrip:10:100:usdc")
	fmt.Println("  clearnode stress-test transfer-roundtrip:10:100:usdc:0.0001")
	fmt.Println("  clearnode stress-test app-session-lifecycle:10:5:3:usdc")
	fmt.Println("  clearnode stress-test app-session-lifecycle:10:5:3:usdc:0.000005")
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
