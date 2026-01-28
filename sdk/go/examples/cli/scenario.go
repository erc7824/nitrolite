package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gopkg.in/yaml.v3"
)

// Scenario represents a sequence of operations to execute
type Scenario struct {
	Name        string    `yaml:"name"`
	Description string    `yaml:"description"`
	Variables   Variables `yaml:"variables,omitempty"`
	Steps       []Step    `yaml:"steps"`
}

// Variables stores key-value pairs that can be referenced in steps
type Variables map[string]string

// Step represents a single operation in a scenario
type Step struct {
	Name        string                 `yaml:"name"`
	Command     string                 `yaml:"command"`
	Args        []string               `yaml:"args,omitempty"`
	WaitBefore  string                 `yaml:"wait_before,omitempty"` // Duration to wait before executing
	WaitAfter   string                 `yaml:"wait_after,omitempty"`  // Duration to wait after executing
	OnError     string                 `yaml:"on_error,omitempty"`    // "stop" or "continue"
	Retry       int                    `yaml:"retry,omitempty"`       // Number of times to retry on failure (default: 0)
	RetryDelay  string                 `yaml:"retry_delay,omitempty"` // Duration to wait between retries (default: 1s)
	Description string                 `yaml:"description,omitempty"`
	Interactive map[string]interface{} `yaml:"interactive,omitempty"` // For interactive commands
}

// ScenarioRunner executes scenarios
type ScenarioRunner struct {
	operator  *Operator
	variables Variables
}

// NewScenarioRunner creates a new scenario runner
func NewScenarioRunner(operator *Operator) *ScenarioRunner {
	return &ScenarioRunner{
		operator:  operator,
		variables: make(Variables),
	}
}

// LoadScenario loads a scenario from a YAML file
func (sr *ScenarioRunner) LoadScenario(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenario file: %w", err)
	}

	var scenario Scenario
	if err := yaml.Unmarshal(data, &scenario); err != nil {
		return nil, fmt.Errorf("failed to parse scenario: %w", err)
	}

	return &scenario, nil
}

// ExecuteScenario runs all steps in a scenario
func (sr *ScenarioRunner) ExecuteScenario(ctx context.Context, scenario *Scenario) error {
	fmt.Println("🎬 Starting Scenario:", scenario.Name)
	if scenario.Description != "" {
		fmt.Println("📝", scenario.Description)
	}
	fmt.Println()

	// Initialize variables
	sr.variables = make(Variables)
	for k, v := range scenario.Variables {
		sr.variables[k] = v
	}

	// Add dynamic variables
	if sr.operator.sdkClient != nil {
		sr.variables["MY_WALLET"] = sr.operator.sdkClient.GetUserAddress()
	}

	startTime := time.Now()
	successCount := 0
	failCount := 0

	for i, step := range scenario.Steps {
		fmt.Printf("▶️  Step %d/%d: %s\n", i+1, len(scenario.Steps), step.Name)
		if step.Description != "" {
			fmt.Printf("   %s\n", step.Description)
		}

		// Wait before executing
		if step.WaitBefore != "" {
			duration, err := time.ParseDuration(step.WaitBefore)
			if err != nil {
				return fmt.Errorf("invalid wait_before duration: %w", err)
			}
			fmt.Printf("   ⏳ Waiting %s...\n", step.WaitBefore)
			time.Sleep(duration)
		}

		// Execute step with retry logic
		maxAttempts := step.Retry + 1 // Retry means additional attempts after first try
		if maxAttempts < 1 {
			maxAttempts = 1
		}

		// Parse retry delay
		retryDelay := time.Second // default 1s
		if step.RetryDelay != "" {
			parsed, err := time.ParseDuration(step.RetryDelay)
			if err != nil {
				return fmt.Errorf("invalid retry_delay duration: %w", err)
			}
			retryDelay = parsed
		}

		var lastErr error
		succeeded := false

		for attempt := 1; attempt <= maxAttempts; attempt++ {
			// Show retry attempt if not first try
			if attempt > 1 {
				fmt.Printf("   🔄 Retry attempt %d/%d (waiting %s)...\n", attempt-1, step.Retry, retryDelay)
				time.Sleep(retryDelay)
			}

			// Execute step
			err := sr.executeStep(ctx, &step)
			if err == nil {
				succeeded = true
				break
			}

			lastErr = err

			// Show failure message
			if attempt < maxAttempts {
				fmt.Printf("   ⚠️  Attempt %d failed: %v\n", attempt, err)
			}
		}

		// Handle final result
		if succeeded {
			successCount++
			fmt.Println("   ✅ Success")
		} else {
			failCount++
			if maxAttempts > 1 {
				fmt.Printf("   ❌ Failed after %d attempts: %v\n", maxAttempts, lastErr)
			} else {
				fmt.Printf("   ❌ Failed: %v\n", lastErr)
			}

			if step.OnError != "continue" {
				fmt.Printf("\n🛑 Scenario stopped at step %d\n", i+1)
				return fmt.Errorf("step %d failed: %w", i+1, lastErr)
			}
			fmt.Println("   ⚠️  Continuing despite error...")
		}

		// Wait after executing (only if succeeded)
		if succeeded && step.WaitAfter != "" {
			duration, err := time.ParseDuration(step.WaitAfter)
			if err != nil {
				return fmt.Errorf("invalid wait_after duration: %w", err)
			}
			fmt.Printf("   ⏳ Waiting %s...\n", step.WaitAfter)
			time.Sleep(duration)
		}

		fmt.Println()
	}

	duration := time.Since(startTime)
	fmt.Println("🎉 Scenario completed!")
	fmt.Printf("📊 Results: %d succeeded, %d failed, %d total\n", successCount, failCount, len(scenario.Steps))
	fmt.Printf("⏱️  Duration: %s\n", duration.Round(time.Millisecond))

	return nil
}

// executeStep executes a single step
func (sr *ScenarioRunner) executeStep(ctx context.Context, step *Step) error {
	// Substitute variables in command and args
	command := sr.substitute(step.Command)
	args := make([]string, len(step.Args))
	for i, arg := range step.Args {
		args[i] = sr.substitute(arg)
	}

	// Handle special commands
	switch command {
	case "set":
		return sr.executeSet(args)
	case "assert-balance":
		return sr.executeAssertBalance(ctx, args)
	case "assert-balance-gt":
		return sr.executeAssertBalanceGreaterThan(ctx, args)
	case "assert-balance-lt":
		return sr.executeAssertBalanceLessThan(ctx, args)
	case "assert-channel-exists":
		return sr.executeAssertChannelExists(ctx, args)
	case "assert-channel-status":
		return sr.executeAssertChannelStatus(ctx, args)
	case "assert-state-version":
		return sr.executeAssertStateVersion(ctx, args)
	case "assert-transaction-count":
		return sr.executeAssertTransactionCount(ctx, args)
	case "wait":
		return sr.executeWait(args)
	case "echo":
		return sr.executeEcho(args)
	default:
		return sr.executeOperatorCommand(ctx, command, args)
	}
}

// executeSet sets a variable
func (sr *ScenarioRunner) executeSet(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("set requires 2 arguments: key value")
	}
	sr.variables[args[0]] = args[1]
	fmt.Printf("   📌 Set %s = %s\n", args[0], args[1])
	return nil
}

// executeAssertBalance checks if a wallet has expected balance
func (sr *ScenarioRunner) executeAssertBalance(ctx context.Context, args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("assert-balance requires 3 arguments: wallet asset expected_amount")
	}

	wallet := args[0]
	asset := args[1]
	expectedStr := args[2]

	expected, err := decimal.NewFromString(expectedStr)
	if err != nil {
		return fmt.Errorf("invalid expected amount: %w", err)
	}

	actual, err := sr.getBalance(ctx, wallet, asset)
	if err != nil {
		return err
	}

	if !actual.Equal(expected) {
		return fmt.Errorf("balance mismatch: expected %s, got %s", expected.String(), actual.String())
	}

	fmt.Printf("   ✓ Balance verified: %s %s == %s\n", actual.String(), asset, expected.String())
	return nil
}

// executeAssertBalanceGreaterThan checks if balance is greater than expected
func (sr *ScenarioRunner) executeAssertBalanceGreaterThan(ctx context.Context, args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("assert-balance-gt requires 3 arguments: wallet asset min_amount")
	}

	wallet := args[0]
	asset := args[1]
	minStr := args[2]

	minAmount, err := decimal.NewFromString(minStr)
	if err != nil {
		return fmt.Errorf("invalid min amount: %w", err)
	}

	actual, err := sr.getBalance(ctx, wallet, asset)
	if err != nil {
		return err
	}

	if !actual.GreaterThan(minAmount) {
		return fmt.Errorf("balance too low: expected > %s, got %s", minAmount.String(), actual.String())
	}

	fmt.Printf("   ✓ Balance verified: %s %s > %s\n", actual.String(), asset, minAmount.String())
	return nil
}

// executeAssertBalanceLessThan checks if balance is less than expected
func (sr *ScenarioRunner) executeAssertBalanceLessThan(ctx context.Context, args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("assert-balance-lt requires 3 arguments: wallet asset max_amount")
	}

	wallet := args[0]
	asset := args[1]
	maxStr := args[2]

	maxAmount, err := decimal.NewFromString(maxStr)
	if err != nil {
		return fmt.Errorf("invalid max amount: %w", err)
	}

	actual, err := sr.getBalance(ctx, wallet, asset)
	if err != nil {
		return err
	}

	if !actual.LessThan(maxAmount) {
		return fmt.Errorf("balance too high: expected < %s, got %s", maxAmount.String(), actual.String())
	}

	fmt.Printf("   ✓ Balance verified: %s %s < %s\n", actual.String(), asset, maxAmount.String())
	return nil
}

// executeAssertChannelExists checks if a channel exists for wallet/asset
func (sr *ScenarioRunner) executeAssertChannelExists(ctx context.Context, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("assert-channel-exists requires 2 arguments: wallet asset")
	}

	wallet := args[0]
	asset := args[1]

	channels, _, err := sr.operator.baseClient.GetChannels(ctx, wallet, nil)
	if err != nil {
		return fmt.Errorf("failed to get channels: %w", err)
	}

	// Get asset info to find matching channels
	assetList, err := sr.operator.baseClient.GetAssets(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to get assets: %w", err)
	}

	// Find token addresses for this asset
	var tokenAddresses []string
	for _, a := range assetList {
		if strings.EqualFold(a.Symbol, asset) {
			for _, token := range a.Tokens {
				tokenAddresses = append(tokenAddresses, strings.ToLower(token.Address))
			}
			break
		}
	}

	if len(tokenAddresses) == 0 {
		return fmt.Errorf("asset %s not found", asset)
	}

	// Check if any channel matches
	for _, channel := range channels {
		for _, tokenAddr := range tokenAddresses {
			if strings.ToLower(channel.TokenAddress) == tokenAddr {
				fmt.Printf("   ✓ Channel exists: %s (status: %v)\n", channel.ChannelID, channel.Status)
				return nil
			}
		}
	}

	return fmt.Errorf("no channel found for wallet %s with asset %s", wallet, asset)
}

// executeAssertChannelStatus checks if a channel has expected status
func (sr *ScenarioRunner) executeAssertChannelStatus(ctx context.Context, args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("assert-channel-status requires 3 arguments: wallet asset expected_status")
	}

	wallet := args[0]
	asset := args[1]
	expectedStatus := strings.ToLower(args[2])

	channels, _, err := sr.operator.baseClient.GetChannels(ctx, wallet, nil)
	if err != nil {
		return fmt.Errorf("failed to get channels: %w", err)
	}

	// Get asset info
	assetList, err := sr.operator.baseClient.GetAssets(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to get assets: %w", err)
	}

	// Find token addresses for this asset
	var tokenAddresses []string
	for _, a := range assetList {
		if strings.EqualFold(a.Symbol, asset) {
			for _, token := range a.Tokens {
				tokenAddresses = append(tokenAddresses, strings.ToLower(token.Address))
			}
			break
		}
	}

	// Find matching channel
	for _, channel := range channels {
		for _, tokenAddr := range tokenAddresses {
			if strings.ToLower(channel.TokenAddress) == tokenAddr {
				actualStatus := strings.ToLower(fmt.Sprintf("%v", channel.Status))

				// Parse expected status
				var matchesExpected bool
				switch expectedStatus {
				case "open", "1":
					matchesExpected = channel.Status == 1 // ChannelStatusOpen
				case "closed", "3":
					matchesExpected = channel.Status == 3 // ChannelStatusClosed
				case "challenged", "2":
					matchesExpected = channel.Status == 2 // ChannelStatusChallenged
				case "void", "0":
					matchesExpected = channel.Status == 0 // ChannelStatusVoid
				default:
					return fmt.Errorf("unknown status: %s (use: open, closed, challenged, void)", expectedStatus)
				}

				if !matchesExpected {
					return fmt.Errorf("channel status mismatch: expected %s, got %s", expectedStatus, actualStatus)
				}

				fmt.Printf("   ✓ Channel status verified: %s (status: %s)\n", channel.ChannelID, expectedStatus)
				return nil
			}
		}
	}

	return fmt.Errorf("no channel found for wallet %s with asset %s", wallet, asset)
}

// executeAssertStateVersion checks if state version matches expected
func (sr *ScenarioRunner) executeAssertStateVersion(ctx context.Context, args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("assert-state-version requires 3 arguments: wallet asset expected_version")
	}

	wallet := args[0]
	asset := args[1]
	expectedVersionStr := args[2]

	expectedVersion, err := strconv.ParseUint(expectedVersionStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid expected version: %w", err)
	}

	state, err := sr.operator.baseClient.GetLatestState(ctx, wallet, asset, false)
	if err != nil {
		return fmt.Errorf("failed to get state: %w", err)
	}

	if state.Version != expectedVersion {
		return fmt.Errorf("state version mismatch: expected %d, got %d", expectedVersion, state.Version)
	}

	fmt.Printf("   ✓ State version verified: %d\n", state.Version)
	return nil
}

// executeAssertTransactionCount checks if transaction count meets criteria
func (sr *ScenarioRunner) executeAssertTransactionCount(ctx context.Context, args []string) error {
	if len(args) < 2 || len(args) > 3 {
		return fmt.Errorf("assert-transaction-count requires 2-3 arguments: wallet min_count [max_count]")
	}

	wallet := args[0]
	minCountStr := args[1]

	minCount, err := strconv.ParseInt(minCountStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid min count: %w", err)
	}

	var maxCount *int64
	if len(args) == 3 {
		parsed, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid max count: %w", err)
		}
		maxCount = &parsed
	}

	// Get transactions
	_, meta, err := sr.operator.baseClient.GetTransactions(ctx, wallet, nil)
	if err != nil {
		return fmt.Errorf("failed to get transactions: %w", err)
	}

	actualCount := int64(meta.TotalCount)

	if actualCount < minCount {
		return fmt.Errorf("transaction count too low: expected >= %d, got %d", minCount, actualCount)
	}

	if maxCount != nil && actualCount > *maxCount {
		return fmt.Errorf("transaction count too high: expected <= %d, got %d", *maxCount, actualCount)
	}

	if maxCount != nil {
		fmt.Printf("   ✓ Transaction count verified: %d (between %d and %d)\n", actualCount, minCount, *maxCount)
	} else {
		fmt.Printf("   ✓ Transaction count verified: %d (>= %d)\n", actualCount, minCount)
	}

	return nil
}

// getBalance is a helper to get balance for a wallet/asset
func (sr *ScenarioRunner) getBalance(ctx context.Context, wallet, asset string) (decimal.Decimal, error) {
	balances, err := sr.operator.baseClient.GetBalances(ctx, wallet)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get balances: %w", err)
	}

	for _, balance := range balances {
		if strings.EqualFold(balance.Asset, asset) {
			return balance.Balance, nil
		}
	}

	return decimal.Zero, fmt.Errorf("asset %s not found in balances", asset)
}

// executeWait waits for a duration
func (sr *ScenarioRunner) executeWait(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("wait requires 1 argument: duration")
	}

	duration, err := time.ParseDuration(args[0])
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	fmt.Printf("   ⏳ Waiting %s...\n", args[0])
	time.Sleep(duration)
	return nil
}

// executeEcho prints a message
func (sr *ScenarioRunner) executeEcho(args []string) error {
	message := strings.Join(args, " ")
	fmt.Printf("   💬 %s\n", message)
	return nil
}

// executeOperatorCommand executes a standard operator command
func (sr *ScenarioRunner) executeOperatorCommand(ctx context.Context, command string, args []string) error {
	// Build full command string
	fullCommand := command
	if len(args) > 0 {
		fullCommand += " " + strings.Join(args, " ")
	}

	fmt.Printf("   🔧 Executing: %s\n", fullCommand)

	// Route to appropriate operator method
	switch command {
	case "ping":
		sr.operator.ping(ctx)
	case "deposit":
		if len(args) < 3 {
			return fmt.Errorf("deposit requires 3 args: chain_id asset amount")
		}
		sr.operator.deposit(ctx, args[0], args[1], args[2])
	case "withdraw":
		if len(args) < 3 {
			return fmt.Errorf("withdraw requires 3 args: chain_id asset amount")
		}
		sr.operator.withdraw(ctx, args[0], args[1], args[2])
	case "transfer":
		if len(args) < 3 {
			return fmt.Errorf("transfer requires 3 args: recipient asset amount")
		}
		sr.operator.transfer(ctx, args[0], args[1], args[2])
	case "balances":
		if len(args) < 1 {
			return fmt.Errorf("balances requires 1 arg: wallet")
		}
		sr.operator.getBalances(ctx, args[0])
	case "channels":
		if len(args) < 1 {
			return fmt.Errorf("channels requires 1 arg: wallet")
		}
		sr.operator.listChannels(ctx, args[0])
	case "transactions":
		if len(args) < 1 {
			return fmt.Errorf("transactions requires 1 arg: wallet")
		}
		sr.operator.listTransactions(ctx, args[0])
	case "state":
		if len(args) < 2 {
			return fmt.Errorf("state requires 2 args: wallet asset")
		}
		sr.operator.getLatestState(ctx, args[0], args[1])
	case "chains":
		sr.operator.listChains(ctx)
	case "assets":
		chainID := ""
		if len(args) > 0 {
			chainID = args[0]
		}
		sr.operator.listAssets(ctx, chainID)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}

	return nil
}

// substitute replaces variables in a string
func (sr *ScenarioRunner) substitute(s string) string {
	result := s
	for k, v := range sr.variables {
		result = strings.ReplaceAll(result, "${"+k+"}", v)
		result = strings.ReplaceAll(result, "$"+k, v)
	}
	return result
}

// SaveScenarioTemplate creates a template scenario file
func SaveScenarioTemplate(path string) error {
	template := Scenario{
		Name:        "Example Scenario",
		Description: "This is an example scenario demonstrating various operations",
		Variables: Variables{
			"CHAIN_ID":  "80002",
			"ASSET":     "usdc",
			"RECIPIENT": "0x1234567890123456789012345678901234567890",
		},
		Steps: []Step{
			{
				Name:        "Check Node",
				Command:     "ping",
				Description: "Verify connection to Clearnode",
			},
			{
				Name:    "List Chains",
				Command: "chains",
			},
			{
				Name:        "Deposit Funds",
				Command:     "deposit",
				Args:        []string{"${CHAIN_ID}", "${ASSET}", "100"},
				WaitAfter:   "2s",
				Description: "Deposit 100 USDC to channel",
			},
			{
				Name:    "Check Balance",
				Command: "balances",
				Args:    []string{"${MY_WALLET}"},
			},
			{
				Name:        "Transfer to Recipient",
				Command:     "transfer",
				Args:        []string{"${RECIPIENT}", "${ASSET}", "25"},
				WaitAfter:   "1s",
				Description: "Send 25 USDC to recipient",
			},
			{
				Name:    "Verify Balance",
				Command: "assert-balance",
				Args:    []string{"${MY_WALLET}", "${ASSET}", "75"},
			},
			{
				Name:    "List Transactions",
				Command: "transactions",
				Args:    []string{"${MY_WALLET}"},
			},
		},
	}

	data, err := yaml.Marshal(&template)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}

	return nil
}
