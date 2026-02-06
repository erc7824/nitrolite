/**
 * Example: Complete App Session Lifecycle
 *
 * This example demonstrates:
 * 1. Create first app session for wallet 1
 * 2. Deposit USDC into first app session by wallet 1
 * 3. Create second app session for wallet 2 with wallet 3 as a participant
 * 4. Deposit WETH into second app session by wallet 2
 * 5. Redistribute app state within app session so that participant with wallet 3 also has some allocation
 * 6. Rebalance 2 app sessions atomically
 * 7. Wallet 3 withdraws from his app session
 * 8. Close both app sessions
 */

import Decimal from 'decimal.js';
import { Client } from '../../src/client';
import { createSigners } from '../../src/signers';
import {
  AppDefinitionV1,
  AppStateUpdateV1,
  AppStateUpdateIntent,
  SignedAppStateUpdateV1,
} from '../../src/app/types';
import { packCreateAppSessionRequestV1, packAppStateUpdateV1 } from '../../src/app/packing';

async function main() {
  // Replace with a real deployment url
  const wsURL = 'wss://deployment.yellow.org/ws';

  // --- 0. Setup Wallets ---
  // Replace these strings with your actual hex private keys
  const wallet1PrivateKey = '0x7d607...';
  const wallet2PrivateKey = '0x9b652...';
  const wallet3PrivateKey = '0xf6369...';

  // Create signers from private keys
  const wallet1Signers = createSigners(wallet1PrivateKey);
  const wallet2Signers = createSigners(wallet2PrivateKey);
  const wallet3Signers = createSigners(wallet3PrivateKey);

  // Extract wallet addresses
  const wallet1Address = wallet1Signers.stateSigner.getAddress(); // 0x053aEAD7d3eebE4359300fDE849bCD9E77384989
  const wallet2Address = wallet2Signers.stateSigner.getAddress(); // 0x2BfA10aAd64Ae0F7855f54f27117Fcc9C61C6770
  const wallet3Address = wallet3Signers.stateSigner.getAddress(); // 0xaB5670b44cb4A3B5535BD637cb600DA572148c98

  console.log('--- Wallets Imported ---');
  console.log(`Wallet 1 Address: ${wallet1Address}`);
  console.log(`Wallet 2 Address: ${wallet2Address}`);
  console.log(`Wallet 3 Address: ${wallet3Address}`);
  console.log('------------------------');

  // Create SDK clients (in a real app, these would be separate instances)
  const wallet1Client = await Client.create(
    wsURL,
    wallet1Signers.stateSigner,
    wallet1Signers.txSigner
  );

  // --- 1. Create App Session 1 (Single Participant: Wallet 1) ---
  console.log('=== Step 1: Creating App Session 1 (Wallet 1 only) ===');

  const session1Definition: AppDefinitionV1 = {
    application: 'test-app',
    participants: [{ walletAddress: wallet1Address, signatureWeight: 100 }],
    quorum: 100,
    nonce: BigInt(Date.now() * 1000000), // Use nanoseconds like Go
  };

  const session1CreateRequest = packCreateAppSessionRequestV1(session1Definition, '{}');
  const wallet1CreateSession1Sig = await wallet1Signers.stateSigner.signMessage(
    session1CreateRequest
  );

  const { appSessionId: session1ID } = await wallet1Client.createAppSession(
    session1Definition,
    '{}',
    [wallet1CreateSession1Sig]
  );
  console.log(`✓ Created App Session 1: ${session1ID}\n`);

  // --- 2. Deposit USDC into Session 1 ---
  console.log('=== Step 2: Depositing USDC into Session 1 ===');

  const session1DepositAmount = new Decimal(0.0001);
  const session1DepositUpdate: AppStateUpdateV1 = {
    appSessionId: session1ID,
    intent: AppStateUpdateIntent.Deposit,
    version: 2n,
    allocations: [
      { participant: wallet1Address, asset: 'usdc', amount: session1DepositAmount },
    ],
    sessionData: '{}',
  };

  const session1DepositRequest = packAppStateUpdateV1(session1DepositUpdate);
  const wallet1DepositSig = await wallet1Signers.stateSigner.signMessage(session1DepositRequest);

  try {
    await wallet1Client.submitAppSessionDeposit(
      session1DepositUpdate,
      [wallet1DepositSig],
      'usdc',
      session1DepositAmount
    );
    console.log(`✓ Deposited ${session1DepositAmount} USDC into Session 1\n`);
  } catch (err) {
    console.log(`⚠ Deposit warning: ${err}`);
  }

  // --- 3. Create App Session 2 (Multi-Party: Wallet 2 & 3) ---
  console.log('=== Step 3: Creating App Session 2 (Wallet 2 & 3) ===');

  const wallet2Client = await Client.create(
    wsURL,
    wallet2Signers.stateSigner,
    wallet2Signers.txSigner
  );

  const session2Definition: AppDefinitionV1 = {
    application: 'multi-party-app',
    participants: [
      { walletAddress: wallet2Address, signatureWeight: 50 },
      { walletAddress: wallet3Address, signatureWeight: 50 },
    ],
    quorum: 100,
    nonce: BigInt(Date.now() * 1000000),
  };

  const session2CreateRequest = packCreateAppSessionRequestV1(session2Definition, '{}');
  const wallet2CreateSession2Sig = await wallet2Signers.stateSigner.signMessage(
    session2CreateRequest
  );
  const wallet3CreateSession2Sig = await wallet3Signers.stateSigner.signMessage(
    session2CreateRequest
  );

  const { appSessionId: session2ID } = await wallet2Client.createAppSession(
    session2Definition,
    '{}',
    [wallet2CreateSession2Sig, wallet3CreateSession2Sig]
  );
  console.log(`✓ Created App Session 2: ${session2ID}\n`);

  // --- 4. Deposit WETH into Session 2 by Wallet 2 ---
  console.log('=== Step 4: Depositing WETH into Session 2 ===');

  const session2DepositAmount = new Decimal(0.015);
  const session2DepositUpdate: AppStateUpdateV1 = {
    appSessionId: session2ID,
    intent: AppStateUpdateIntent.Deposit,
    version: 2n,
    allocations: [
      { participant: wallet2Address, asset: 'weth', amount: session2DepositAmount },
    ],
    sessionData: '{}',
  };

  const session2DepositRequest = packAppStateUpdateV1(session2DepositUpdate);
  const wallet2DepositSig = await wallet2Signers.stateSigner.signMessage(session2DepositRequest);
  const wallet3DepositSig = await wallet3Signers.stateSigner.signMessage(session2DepositRequest);

  const nodeSig = await wallet2Client.submitAppSessionDeposit(
    session2DepositUpdate,
    [wallet2DepositSig, wallet3DepositSig],
    'weth',
    session2DepositAmount
  );
  console.log(`✓ Deposited ${session2DepositAmount} WETH into Session 2 (Node Sig: ${nodeSig})\n`);

  // Check Session 2 state before redistribution
  const { sessions: session2InfoBeforeRedist } = await wallet2Client.getAppSessions({
    appSessionId: session2ID,
  });
  if (session2InfoBeforeRedist.length > 0) {
    console.log(
      `Session 2 before redistribution - Version: ${session2InfoBeforeRedist[0].version}, Allocations: ${JSON.stringify(session2InfoBeforeRedist[0].allocations)}\n`
    );
  }

  // --- 5. Redistribute within Session 2 (Wallet 2 -> Wallet 3) ---
  console.log('=== Step 5: Redistributing funds in Session 2 ===');

  const session2RedistributeUpdate: AppStateUpdateV1 = {
    appSessionId: session2ID,
    intent: AppStateUpdateIntent.Operate,
    version: 3n,
    allocations: [
      { participant: wallet2Address, asset: 'weth', amount: new Decimal(0.01) },
      { participant: wallet3Address, asset: 'weth', amount: new Decimal(0.005) },
    ],
    sessionData: '{}',
  };

  const session2RedistributeRequest = packAppStateUpdateV1(session2RedistributeUpdate);
  const wallet2RedistributeSig = await wallet2Signers.stateSigner.signMessage(
    session2RedistributeRequest
  );
  const wallet3RedistributeSig = await wallet3Signers.stateSigner.signMessage(
    session2RedistributeRequest
  );

  // Multi-sig required for state transition
  try {
    await wallet2Client.submitAppState(session2RedistributeUpdate, [
      wallet2RedistributeSig,
      wallet3RedistributeSig,
    ]);
    console.log('✓ Redistributed WETH: Wallet 2 (0.01) -> Wallet 3 (0.005)\n');
  } catch (err) {
    console.error(`Redistribution failed: ${err}`);
    throw err;
  }

  // --- 6. Rebalance Both App Sessions Atomically ---
  console.log('=== Step 6: Atomic Rebalance Across Sessions ===');

  // Check current allocations before rebalance
  const { sessions: session1InfoBeforeRebalance } = await wallet1Client.getAppSessions({
    appSessionId: session1ID,
  });
  if (session1InfoBeforeRebalance.length > 0) {
    console.log(
      `Session 1 before rebalance - Version: ${session1InfoBeforeRebalance[0].version}, Allocations: ${JSON.stringify(session1InfoBeforeRebalance[0].allocations)}`
    );
  }

  const { sessions: session2InfoBeforeRebalance } = await wallet2Client.getAppSessions({
    appSessionId: session2ID,
  });
  if (session2InfoBeforeRebalance.length > 0) {
    console.log(
      `Session 2 before rebalance - Version: ${session2InfoBeforeRebalance[0].version}, Allocations: ${JSON.stringify(session2InfoBeforeRebalance[0].allocations)}\n`
    );
  }

  // Prepare rebalance updates for both sessions
  const session1RebalanceUpdate: AppStateUpdateV1 = {
    appSessionId: session1ID,
    intent: AppStateUpdateIntent.Rebalance,
    version: 3n,
    allocations: [
      { participant: wallet1Address, asset: 'weth', amount: new Decimal(0.005) },
      { participant: wallet1Address, asset: 'usdc', amount: new Decimal(0.00005) },
    ],
    sessionData: '{}',
  };

  const session1RebalanceRequest = packAppStateUpdateV1(session1RebalanceUpdate);
  const wallet1RebalanceSig = await wallet1Signers.stateSigner.signMessage(
    session1RebalanceRequest
  );

  const session2RebalanceUpdate: AppStateUpdateV1 = {
    appSessionId: session2ID,
    intent: AppStateUpdateIntent.Rebalance,
    version: 4n,
    allocations: [
      { participant: wallet2Address, asset: 'usdc', amount: new Decimal(0.00005) },
      { participant: wallet2Address, asset: 'weth', amount: new Decimal(0.005) },
      { participant: wallet3Address, asset: 'weth', amount: new Decimal(0.005) },
    ],
    sessionData: '{}',
  };

  const session2RebalanceRequest = packAppStateUpdateV1(session2RebalanceUpdate);
  const wallet2RebalanceSig = await wallet2Signers.stateSigner.signMessage(
    session2RebalanceRequest
  );
  const wallet3RebalanceSig = await wallet3Signers.stateSigner.signMessage(
    session2RebalanceRequest
  );

  // Submit atomic rebalance
  const signedRebalanceUpdates: SignedAppStateUpdateV1[] = [
    {
      appStateUpdate: session1RebalanceUpdate,
      quorumSigs: [wallet1RebalanceSig],
    },
    {
      appStateUpdate: session2RebalanceUpdate,
      quorumSigs: [wallet2RebalanceSig, wallet3RebalanceSig],
    },
  ];

  try {
    const rebalanceBatchID = await wallet2Client.rebalanceAppSessions(signedRebalanceUpdates);
    console.log(`✓ Atomic Rebalance Submitted. BatchID: ${rebalanceBatchID}\n`);
  } catch (err) {
    console.log(`⚠ Rebalance Error: ${err}\n`);
  }

  // --- 7. Wallet 3 Withdraws from Session 2 ---
  console.log('=== Step 7: Wallet 3 Withdrawing from Session 2 ===');

  const session2WithdrawUpdate: AppStateUpdateV1 = {
    appSessionId: session2ID,
    intent: AppStateUpdateIntent.Withdraw,
    version: 5n,
    allocations: [
      { participant: wallet2Address, asset: 'usdc', amount: new Decimal(0.00005) },
      { participant: wallet2Address, asset: 'weth', amount: new Decimal(0.005) },
      { participant: wallet3Address, asset: 'weth', amount: new Decimal(0.001) },
    ],
    sessionData: '{}',
  };

  const session2WithdrawRequest = packAppStateUpdateV1(session2WithdrawUpdate);
  const wallet2WithdrawSig = await wallet2Signers.stateSigner.signMessage(
    session2WithdrawRequest
  );
  const wallet3WithdrawSig = await wallet3Signers.stateSigner.signMessage(
    session2WithdrawRequest
  );

  try {
    await wallet2Client.submitAppState(session2WithdrawUpdate, [
      wallet2WithdrawSig,
      wallet3WithdrawSig,
    ]);
    console.log('✓ Wallet 3 successfully withdrew 0.004 WETH back to channel\n');
  } catch (err) {
    console.log(`⚠ Withdraw Error: ${err}\n`);
  }

  // --- 8. Close Both App Sessions ---
  console.log('=== Step 8: Closing Both App Sessions ===');

  // Close Session 1
  const session1CloseUpdate: AppStateUpdateV1 = {
    appSessionId: session1ID,
    intent: AppStateUpdateIntent.Close,
    version: 4n,
    allocations: [
      { participant: wallet1Address, asset: 'weth', amount: new Decimal(0.005) },
      { participant: wallet1Address, asset: 'usdc', amount: new Decimal(0.00005) },
    ],
    sessionData: '{}',
  };

  const session1CloseRequest = packAppStateUpdateV1(session1CloseUpdate);
  const wallet1CloseSig = await wallet1Signers.stateSigner.signMessage(session1CloseRequest);

  try {
    await wallet1Client.submitAppState(session1CloseUpdate, [wallet1CloseSig]);
    console.log('✓ Session 1 successfully closed');
  } catch (err) {
    console.log(`⚠ Close Session 1 Error: ${err}`);
  }

  // Close Session 2
  const session2CloseUpdate: AppStateUpdateV1 = {
    appSessionId: session2ID,
    intent: AppStateUpdateIntent.Close,
    version: 6n,
    allocations: [
      { participant: wallet2Address, asset: 'usdc', amount: new Decimal(0.00005) },
      { participant: wallet2Address, asset: 'weth', amount: new Decimal(0.005) },
      { participant: wallet3Address, asset: 'weth', amount: new Decimal(0.001) },
    ],
    sessionData: '{}',
  };

  const session2CloseRequest = packAppStateUpdateV1(session2CloseUpdate);
  const wallet2CloseSig = await wallet2Signers.stateSigner.signMessage(session2CloseRequest);
  const wallet3CloseSig = await wallet3Signers.stateSigner.signMessage(session2CloseRequest);

  try {
    await wallet2Client.submitAppState(session2CloseUpdate, [
      wallet2CloseSig,
      wallet3CloseSig,
    ]);
    console.log('✓ Session 2 successfully closed');
  } catch (err) {
    console.log(`⚠ Close Session 2 Error: ${err}`);
  }

  console.log('\n=== Example Complete ===');

  // Close clients
  await wallet1Client.close();
  await wallet2Client.close();

  // Exit successfully
  process.exit(0);
}

// Run the example
main().catch((error) => {
  console.error('Fatal error:', error);
  process.exit(1);
});
