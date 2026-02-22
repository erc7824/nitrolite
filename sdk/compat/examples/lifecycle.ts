/**
 * Compat Layer Lifecycle Example
 *
 * Demonstrates using @erc7824/nitrolite-compat to interact with a clearnode.
 * Exercises: ping, config, assets, balances, channels, app sessions, transfers.
 *
 * Prerequisites:
 *   - A running clearnode (local or remote)
 *   - PRIVATE_KEY env var set to an Ethereum private key
 *   - CLEARNODE_WS_URL env var (defaults to ws://localhost:7824/ws)
 *   - CHAIN_ID env var (defaults to 11155111 for Sepolia)
 *
 * Usage:
 *   export PRIVATE_KEY=0x...
 *   export CLEARNODE_WS_URL=ws://localhost:7824/ws
 *   npx tsx examples/lifecycle.ts
 */

import { NitroliteClient, blockchainRPCsFromEnv } from '@erc7824/nitrolite-compat';
import { createWalletClient, http } from 'viem';
import { sepolia } from 'viem/chains';
import { privateKeyToAccount } from 'viem/accounts';

const PRIVATE_KEY = process.env.PRIVATE_KEY as `0x${string}`;
const WS_URL = process.env.CLEARNODE_WS_URL || 'ws://localhost:7824/ws';
const CHAIN_ID = Number(process.env.CHAIN_ID || '11155111');

if (!PRIVATE_KEY) {
    console.error('Set PRIVATE_KEY env var (e.g. export PRIVATE_KEY=0x...)');
    process.exit(1);
}

let passed = 0;
let failed = 0;
function ok(label: string, detail?: string) {
    passed++;
    console.log(`  PASS  ${label}${detail ? ' -- ' + detail : ''}`);
}
function fail(label: string, err: any) {
    failed++;
    console.error(`  FAIL  ${label} -- ${err?.message ?? err}`);
}

async function main() {
    console.log('=== Compat Layer Lifecycle Example ===\n');

    const account = privateKeyToAccount(PRIVATE_KEY);
    console.log(`Wallet: ${account.address}`);
    console.log(`Clearnode: ${WS_URL}`);
    console.log(`Chain: ${CHAIN_ID}\n`);

    const walletClient = createWalletClient({
        chain: sepolia,
        transport: http(process.env.RPC_URL || 'https://1rpc.io/sepolia'),
        account,
    });

    // ================================================================
    // 1. Create compat client
    // ================================================================
    console.log('-- Initializing compat client --');
    let client: NitroliteClient;
    try {
        client = await NitroliteClient.create({
            wsURL: WS_URL,
            walletClient,
            chainId: CHAIN_ID,
            blockchainRPCs: blockchainRPCsFromEnv(),
        });
        ok('NitroliteClient.create()');
    } catch (err: any) {
        fail('NitroliteClient.create()', err);
        process.exit(1);
    }

    // ================================================================
    // 2. Node queries
    // ================================================================
    console.log('\n-- Node Queries --');

    try {
        await client.ping();
        ok('ping()');
    } catch (e: any) { fail('ping()', e); }

    try {
        const config = await client.getConfig();
        ok('getConfig()', `blockchains=${config.blockchains?.length ?? '?'}`);
    } catch (e: any) { fail('getConfig()', e); }

    try {
        const assets = await client.getAssetsList();
        ok('getAssetsList()', `${assets.length} asset(s): ${assets.map(a => a.symbol).join(', ')}`);
    } catch (e: any) { fail('getAssetsList()', e); }

    // ================================================================
    // 3. Balance & ledger queries
    // ================================================================
    console.log('\n-- Balance & Ledger --');

    try {
        const balances = await client.getBalances();
        ok('getBalances()', `${balances.length} balance(s)`);
    } catch (e: any) { fail('getBalances()', e); }

    try {
        const entries = await client.getLedgerEntries();
        ok('getLedgerEntries()', `${entries.length} entry/entries`);
    } catch (e: any) { fail('getLedgerEntries()', e); }

    // ================================================================
    // 4. Channel queries
    // ================================================================
    console.log('\n-- Channels --');

    try {
        const channels = await client.getChannels();
        ok('getChannels()', `${channels.length} channel(s)`);
    } catch (e: any) { fail('getChannels()', e); }

    try {
        const info = await client.getAccountInfo();
        ok('getAccountInfo()', `available=${info.available}, channels=${info.channelCount}`);
    } catch (e: any) { fail('getAccountInfo()', e); }

    // ================================================================
    // 5. App sessions
    // ================================================================
    console.log('\n-- App Sessions --');

    try {
        const sessions = await client.getAppSessionsList();
        ok('getAppSessionsList()', `${sessions.length} session(s)`);
    } catch (e: any) { fail('getAppSessionsList()', e); }

    try {
        const openSessions = await client.getAppSessionsList(undefined, 'open');
        ok('getAppSessionsList(status=open)', `${openSessions.length} open`);
    } catch (e: any) { fail('getAppSessionsList(status=open)', e); }

    try {
        const closedSessions = await client.getAppSessionsList(undefined, 'closed');
        ok('getAppSessionsList(status=closed)', `${closedSessions.length} closed`);
    } catch (e: any) { fail('getAppSessionsList(status=closed)', e); }

    // Try getAppDefinition on an existing session
    try {
        const sessions = await client.getAppSessionsList();
        if (sessions.length > 0) {
            const def = await client.getAppDefinition(sessions[0].app_session_id);
            ok('getAppDefinition()', `app=${def.protocol}, quorum=${def.quorum}`);
        } else {
            ok('getAppDefinition() (skipped)', 'no sessions to query');
        }
    } catch (e: any) { fail('getAppDefinition()', e); }

    // ================================================================
    // 6. Asset resolution helpers
    // ================================================================
    console.log('\n-- Asset Helpers --');

    try {
        const assets = await client.getAssetsList();
        if (assets.length > 0) {
            const token = assets[0];
            const resolved = client.resolveToken(token.token);
            ok('resolveToken()', `${token.token.slice(0, 10)}... -> ${resolved.symbol}`);

            const bySymbol = client.resolveAsset(resolved.symbol);
            ok('resolveAsset()', `${resolved.symbol} -> decimals=${bySymbol.decimals}`);

            const decimals = client.getTokenDecimals(token.token);
            ok('getTokenDecimals()', `${decimals}`);

            const formatted = client.formatAmount(token.token, 1000000n);
            ok('formatAmount()', `1000000 raw -> ${formatted}`);

            const parsed = client.parseAmount(token.token, '1.0');
            ok('parseAmount()', `1.0 -> ${parsed} raw`);

            const display = client.resolveAssetDisplay(token.token);
            ok('resolveAssetDisplay()', display ? `${display.symbol} (${display.decimals} dec)` : 'null');

            const channel = client.findOpenChannel(token.token);
            ok('findOpenChannel()', channel ? `found: ${channel.channel_id.slice(0, 16)}...` : 'none');
        }
    } catch (e: any) { fail('asset helpers', e); }

    // ================================================================
    // 7. Cleanup
    // ================================================================
    console.log('\n-- Cleanup --');

    try {
        await client.close();
        ok('close()');
    } catch (e: any) { fail('close()', e); }

    // ================================================================
    // Summary
    // ================================================================
    console.log(`\n=== Results: ${passed} passed, ${failed} failed ===`);
    if (failed > 0) process.exit(1);
}

main().catch((err) => {
    console.error('Fatal:', err);
    process.exit(1);
});
