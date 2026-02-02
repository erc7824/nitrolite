package main

// Example: Complete App Session Lifecycle
//
// This example demonstrates:
// 1. Create first app session for wallet 1
// 2. Deposit USDC into first app session by wallet 1
// 3. Create second app session for wallet 2 with wallet 3 as a participant
// 4. Deposit WETH into second app session by wallet 2
// 5. Redistribute app state within app session so that participant with wallet 3 also has some allocation
// 6. Rebalance 2 app sessions atomically
// 7. Wallet 3 withdraws from his app session
// 8. Close both app sessions

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/sign"
	sdk "github.com/erc7824/nitrolite/sdk/go"
	"github.com/shopspring/decimal"
)

func main() {
	ctx := context.Background()
	wsURL := "wss://clearnode-v1-rc.yellow.org/ws"

	// --- 0. Setup Wallets ---
	// Replace these strings with your actual hex private keys
	wallet1PrivateKey := "0x7d6071201765d2630ca9eb83cbe3e2e2e76f9b56ea3ed13a49a00208ebcdf843"
	wallet2PrivateKey := "0x9b6521133af49807e72b8ecc68ef79706fe374685214130079c375810ec47fe3"
	wallet3PrivateKey := "0xf636952f9d68984a78ef45ea82480723b8a2c40127111cf83d384f8dcd3b77f8"

	// Create signers from private keys
	wallet1Signer, err := sign.NewEthereumMsgSigner(wallet1PrivateKey)
	if err != nil {
		log.Fatalf("Invalid wallet 1 private key: %v", err)
	}
	wallet2Signer, err := sign.NewEthereumMsgSigner(wallet2PrivateKey)
	if err != nil {
		log.Fatalf("Invalid wallet 2 private key: %v", err)
	}
	wallet3Signer, err := sign.NewEthereumMsgSigner(wallet3PrivateKey)
	if err != nil {
		log.Fatalf("Invalid wallet 3 private key: %v", err)
	}

	// Extract wallet addresses
	wallet1Address := wallet1Signer.PublicKey().Address().String() // 0x053aEAD7d3eebE4359300fDE849bCD9E77384989
	wallet2Address := wallet2Signer.PublicKey().Address().String() // 0x2BfA10aAd64Ae0F7855f54f27117Fcc9C61C6770
	wallet3Address := wallet3Signer.PublicKey().Address().String() // 0xaB5670b44cb4A3B5535BD637cb600DA572148c98

	fmt.Println("--- Wallets Imported ---")
	fmt.Printf("Wallet 1 Address: %s\n", wallet1Address)
	fmt.Printf("Wallet 2 Address: %s\n", wallet2Address)
	fmt.Printf("Wallet 3 Address: %s\n", wallet3Address)
	fmt.Println("------------------------")

	// Create SDK clients (in a real app, these would be separate instances)
	wallet1Client, err := sdk.NewClient(wsURL, wallet1Signer, wallet1Signer)
	if err != nil {
		log.Fatal(err)
	}

	// --- 1. Create App Session 1 (Single Participant: Wallet 1) ---
	fmt.Println("=== Step 1: Creating App Session 1 (Wallet 1 only) ===")

	session1Definition := app.AppDefinitionV1{
		Application: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: wallet1Address, SignatureWeight: 100},
		},
		Quorum: 100,
		Nonce:  uint64(time.Now().UnixNano()),
	}

	session1CreateRequest, err := app.PackCreateAppSessionRequestV1(session1Definition, "{}")
	if err != nil {
		log.Fatal(err)
	}

	wallet1CreateSession1Sig, _ := wallet1Signer.Sign(session1CreateRequest)
	session1ID, _, _, err := wallet1Client.CreateAppSession(ctx, session1Definition, "{}", []string{wallet1CreateSession1Sig.String()})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Created App Session 1: %s\n\n", session1ID)

	// --- 2. Deposit USDC into Session 1 ---
	fmt.Println("=== Step 2: Depositing USDC into Session 1 ===")

	session1DepositAmount := decimal.NewFromFloat(0.0001)
	session1DepositUpdate := app.AppStateUpdateV1{
		AppSessionID: session1ID,
		Intent:       app.AppStateUpdateIntentDeposit,
		Version:      2,
		Allocations:  []app.AppAllocationV1{{Participant: wallet1Address, Asset: "usdc", Amount: session1DepositAmount}},
	}

	session1DepositRequest, err := app.PackAppStateUpdateV1(session1DepositUpdate)
	if err != nil {
		log.Fatal(err)
	}

	wallet1DepositSig, _ := wallet1Signer.Sign(session1DepositRequest)

	// Build channel state for deposit
	wallet1USDCState, err := wallet1Client.GetLatestState(ctx, wallet1Address, "usdc", false)
	if err != nil {
		log.Fatal(err)
	}

	wallet1USDCNextState := wallet1USDCState.NextState()

	_, err = wallet1USDCNextState.ApplyCommitTransition(session1ID, session1DepositAmount)
	if err != nil {
		log.Fatal(err)
	}

	wallet1USDCStateSig, err := wallet1Client.SignState(wallet1USDCNextState)
	if err != nil {
		log.Fatal(err)
	}
	wallet1USDCNextState.UserSig = &wallet1USDCStateSig

	_, err = wallet1Client.SubmitAppSessionDeposit(ctx, session1DepositUpdate, []string{wallet1DepositSig.String()}, *wallet1USDCNextState)
	if err != nil {
		log.Printf("⚠ Deposit warning: %v", err)
	}
	fmt.Printf("✓ Deposited %s USDC into Session 1\n\n", session1DepositAmount)

	// --- 3. Create App Session 2 (Multi-Party: Wallet 2 & 3) ---
	fmt.Println("=== Step 3: Creating App Session 2 (Wallet 2 & 3) ===")

	wallet2Client, err := sdk.NewClient(wsURL, wallet2Signer, wallet2Signer)
	if err != nil {
		log.Fatal(err)
	}

	session2Definition := app.AppDefinitionV1{
		Application: "multi-party-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: wallet2Address, SignatureWeight: 50},
			{WalletAddress: wallet3Address, SignatureWeight: 50},
		},
		Quorum: 100,
		Nonce:  uint64(time.Now().UnixNano()),
	}

	session2CreateRequest, err := app.PackCreateAppSessionRequestV1(session2Definition, "{}")
	if err != nil {
		log.Fatal(err)
	}

	wallet2CreateSession2Sig, _ := wallet2Signer.Sign(session2CreateRequest)
	wallet3CreateSession2Sig, _ := wallet3Signer.Sign(session2CreateRequest)
	session2ID, _, _, err := wallet2Client.CreateAppSession(ctx, session2Definition, "{}", []string{wallet2CreateSession2Sig.String(), wallet3CreateSession2Sig.String()})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Created App Session 2: %s\n\n", session2ID)

	// --- 4. Deposit WETH into Session 2 by Wallet 2 ---
	fmt.Println("=== Step 4: Depositing WETH into Session 2 ===")

	session2DepositAmount := decimal.NewFromFloat(0.015)
	session2DepositUpdate := app.AppStateUpdateV1{
		AppSessionID: session2ID,
		Intent:       app.AppStateUpdateIntentDeposit,
		Version:      2,
		Allocations:  []app.AppAllocationV1{{Participant: wallet2Address, Asset: "weth", Amount: session2DepositAmount}},
	}

	session2DepositRequest, err := app.PackAppStateUpdateV1(session2DepositUpdate)
	if err != nil {
		log.Fatal(err)
	}

	wallet2DepositSig, _ := wallet2Signer.Sign(session2DepositRequest)
	wallet3DepositSig, _ := wallet3Signer.Sign(session2DepositRequest)

	// Build channel state for deposit
	wallet2WETHState, err := wallet2Client.GetLatestState(ctx, wallet2Address, "weth", false)
	if err != nil {
		log.Fatal(err)
	}

	wallet2WETHNextState := wallet2WETHState.NextState()

	_, err = wallet2WETHNextState.ApplyCommitTransition(session2ID, session2DepositAmount)
	if err != nil {
		log.Fatal(err)
	}

	wallet2WETHStateSig, err := wallet2Client.SignState(wallet2WETHNextState)
	if err != nil {
		log.Fatal(err)
	}
	wallet2WETHNextState.UserSig = &wallet2WETHStateSig

	nodeSig, err := wallet2Client.SubmitAppSessionDeposit(ctx, session2DepositUpdate, []string{wallet2DepositSig.String(), wallet3DepositSig.String()}, *wallet2WETHNextState)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Deposited %s WETH into Session 2 (Node Sig: %s)\n\n", session2DepositAmount, nodeSig)

	// Check Session 2 state before redistribution
	session2InfoBeforeRedist, _, err := wallet2Client.GetAppSessions(ctx, &sdk.GetAppSessionsOptions{AppSessionID: &session2ID})
	if err != nil {
		log.Fatal(err)
	}
	if len(session2InfoBeforeRedist) > 0 {
		fmt.Printf("Session 2 before redistribution - Version: %d, Allocations: %+v\n\n", session2InfoBeforeRedist[0].Version, session2InfoBeforeRedist[0].Allocations)
	}

	// --- 5. Redistribute within Session 2 (Wallet 2 -> Wallet 3) ---
	fmt.Println("=== Step 5: Redistributing funds in Session 2 ===")

	session2RedistributeUpdate := app.AppStateUpdateV1{
		AppSessionID: session2ID,
		Intent:       app.AppStateUpdateIntentOperate,
		Version:      3,
		Allocations: []app.AppAllocationV1{
			{Participant: wallet2Address, Asset: "weth", Amount: decimal.NewFromFloat(0.01)},
			{Participant: wallet3Address, Asset: "weth", Amount: decimal.NewFromFloat(0.005)},
		},
	}

	session2RedistributeRequest, err := app.PackAppStateUpdateV1(session2RedistributeUpdate)
	if err != nil {
		log.Fatal(err)
	}

	wallet2RedistributeSig, _ := wallet2Signer.Sign(session2RedistributeRequest)
	wallet3RedistributeSig, _ := wallet3Signer.Sign(session2RedistributeRequest)

	// Multi-sig required for state transition
	err = wallet2Client.SubmitAppState(ctx, session2RedistributeUpdate, []string{wallet2RedistributeSig.String(), wallet3RedistributeSig.String()})
	if err != nil {
		log.Fatalf("Redistribution failed: %v", err)
	}
	fmt.Println("✓ Redistributed WETH: Wallet 2 (0.01) -> Wallet 3 (0.005)")

	// --- 6. Rebalance Both App Sessions Atomically ---
	fmt.Println("=== Step 6: Atomic Rebalance Across Sessions ===")

	// Check current allocations before rebalance
	session1InfoBeforeRebalance, _, err := wallet1Client.GetAppSessions(ctx, &sdk.GetAppSessionsOptions{AppSessionID: &session1ID})
	if err != nil {
		log.Fatal(err)
	}
	if len(session1InfoBeforeRebalance) > 0 {
		fmt.Printf("Session 1 before rebalance - Version: %d, Allocations: %+v\n", session1InfoBeforeRebalance[0].Version, session1InfoBeforeRebalance[0].Allocations)
	}

	session2InfoBeforeRebalance, _, err := wallet2Client.GetAppSessions(ctx, &sdk.GetAppSessionsOptions{AppSessionID: &session2ID})
	if err != nil {
		log.Fatal(err)
	}
	if len(session2InfoBeforeRebalance) > 0 {
		fmt.Printf("Session 2 before rebalance - Version: %d, Allocations: %+v\n\n", session2InfoBeforeRebalance[0].Version, session2InfoBeforeRebalance[0].Allocations)
	}

	// Prepare rebalance updates for both sessions
	session1RebalanceUpdate := app.AppStateUpdateV1{
		AppSessionID: session1ID,
		Intent:       app.AppStateUpdateIntentRebalance,
		Version:      3,
		Allocations: []app.AppAllocationV1{
			{Participant: wallet1Address, Asset: "weth", Amount: decimal.NewFromFloat(0.005)},
			{Participant: wallet1Address, Asset: "usdc", Amount: decimal.NewFromFloat(0.00005)},
		},
	}

	session1RebalanceRequest, err := app.PackAppStateUpdateV1(session1RebalanceUpdate)
	if err != nil {
		log.Fatal(err)
	}

	wallet1RebalanceSig, _ := wallet1Signer.Sign(session1RebalanceRequest)

	session2RebalanceUpdate := app.AppStateUpdateV1{
		AppSessionID: session2ID,
		Intent:       app.AppStateUpdateIntentRebalance,
		Version:      4,
		Allocations: []app.AppAllocationV1{
			{Participant: wallet2Address, Asset: "usdc", Amount: decimal.NewFromFloat(0.00005)},
			{Participant: wallet2Address, Asset: "weth", Amount: decimal.NewFromFloat(0.005)},
			{Participant: wallet3Address, Asset: "weth", Amount: decimal.NewFromFloat(0.005)},
		},
	}

	session2RebalanceRequest, err := app.PackAppStateUpdateV1(session2RebalanceUpdate)
	if err != nil {
		log.Fatal(err)
	}

	wallet2RebalanceSig, _ := wallet2Signer.Sign(session2RebalanceRequest)
	wallet3RebalanceSig, _ := wallet3Signer.Sign(session2RebalanceRequest)

	// Submit atomic rebalance
	signedRebalanceUpdates := []app.SignedAppStateUpdateV1{
		{
			AppStateUpdate: session1RebalanceUpdate,
			QuorumSigs:     []string{wallet1RebalanceSig.String()},
		},
		{
			AppStateUpdate: session2RebalanceUpdate,
			QuorumSigs:     []string{wallet2RebalanceSig.String(), wallet3RebalanceSig.String()},
		},
	}

	rebalanceBatchID, err := wallet2Client.RebalanceAppSessions(ctx, signedRebalanceUpdates)
	if err != nil {
		log.Printf("⚠ Rebalance Error: %v", err)
	} else {
		fmt.Printf("✓ Atomic Rebalance Submitted. BatchID: %s\n\n", rebalanceBatchID)
	}

	// --- 7. Wallet 3 Withdraws from Session 2 ---
	fmt.Println("=== Step 7: Wallet 3 Withdrawing from Session 2 ===")

	session2WithdrawUpdate := app.AppStateUpdateV1{
		AppSessionID: session2ID,
		Intent:       app.AppStateUpdateIntentWithdraw,
		Version:      5,
		Allocations: []app.AppAllocationV1{
			{Participant: wallet2Address, Asset: "usdc", Amount: decimal.NewFromFloat(0.00005)},
			{Participant: wallet2Address, Asset: "weth", Amount: decimal.NewFromFloat(0.005)},
			{Participant: wallet3Address, Asset: "weth", Amount: decimal.NewFromFloat(0.001)},
		},
	}

	session2WithdrawRequest, err := app.PackAppStateUpdateV1(session2WithdrawUpdate)
	if err != nil {
		log.Fatal(err)
	}

	wallet2WithdrawSig, _ := wallet2Signer.Sign(session2WithdrawRequest)
	wallet3WithdrawSig, _ := wallet3Signer.Sign(session2WithdrawRequest)

	err = wallet2Client.SubmitAppState(ctx, session2WithdrawUpdate, []string{wallet2WithdrawSig.String(), wallet3WithdrawSig.String()})
	if err != nil {
		log.Printf("⚠ Withdraw Error: %v", err)
	} else {
		fmt.Println("✓ Wallet 3 successfully withdrew 0.004 WETH back to channel")
	}

	// --- 8. Close Both App Sessions ---
	fmt.Println("=== Step 8: Closing Both App Sessions ===")

	// Close Session 1
	session1CloseUpdate := app.AppStateUpdateV1{
		AppSessionID: session1ID,
		Intent:       app.AppStateUpdateIntentClose,
		Version:      4,
		Allocations: []app.AppAllocationV1{
			{Participant: wallet1Address, Asset: "weth", Amount: decimal.NewFromFloat(0.005)},
			{Participant: wallet1Address, Asset: "usdc", Amount: decimal.NewFromFloat(0.00005)},
		},
	}

	session1CloseRequest, err := app.PackAppStateUpdateV1(session1CloseUpdate)
	if err != nil {
		log.Fatal(err)
	}

	wallet1CloseSig, _ := wallet1Signer.Sign(session1CloseRequest)

	err = wallet1Client.SubmitAppState(ctx, session1CloseUpdate, []string{wallet1CloseSig.String()})
	if err != nil {
		log.Printf("⚠ Close Session 1 Error: %v", err)
	} else {
		fmt.Println("✓ Session 1 successfully closed")
	}

	// Close Session 2
	session2CloseUpdate := app.AppStateUpdateV1{
		AppSessionID: session2ID,
		Intent:       app.AppStateUpdateIntentClose,
		Version:      6,
		Allocations: []app.AppAllocationV1{
			{Participant: wallet2Address, Asset: "usdc", Amount: decimal.NewFromFloat(0.00005)},
			{Participant: wallet2Address, Asset: "weth", Amount: decimal.NewFromFloat(0.005)},
			{Participant: wallet3Address, Asset: "weth", Amount: decimal.NewFromFloat(0.001)},
		},
	}

	session2CloseRequest, err := app.PackAppStateUpdateV1(session2CloseUpdate)
	if err != nil {
		log.Fatal(err)
	}

	wallet2CloseSig, _ := wallet2Signer.Sign(session2CloseRequest)
	wallet3CloseSig, _ := wallet3Signer.Sign(session2CloseRequest)

	err = wallet2Client.SubmitAppState(ctx, session2CloseUpdate, []string{wallet2CloseSig.String(), wallet3CloseSig.String()})
	if err != nil {
		log.Printf("⚠ Close Session 2 Error: %v", err)
	} else {
		fmt.Println("✓ Session 2 successfully closed")
	}

	fmt.Println("\n=== Example Complete ===")
}
