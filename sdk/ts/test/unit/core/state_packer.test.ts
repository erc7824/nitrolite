import Decimal from 'decimal.js';
import { Address } from 'viem';
import { newStatePackerV1 } from '../../../src/core/state_packer';
import { State, TransitionType, newTransition, Ledger } from '../../../src/core/types';
import { AssetStore } from '../../../src/core/interface';

// Mock asset store for testing
class MockAssetStore implements AssetStore {
  private tokens: Map<string, number> = new Map();

  addToken(blockchainId: bigint, tokenAddress: Address, decimals: number): void {
    const key = `${blockchainId}-${tokenAddress.toLowerCase()}`;
    this.tokens.set(key, decimals);
  }

  async getAssetDecimals(asset: string): Promise<number> {
    return 0; // Not used in tests
  }

  async getTokenDecimals(blockchainId: bigint, tokenAddress: Address): Promise<number> {
    const key = `${blockchainId}-${tokenAddress.toLowerCase()}`;
    const decimals = this.tokens.get(key);
    if (decimals === undefined) {
      throw new Error(`Token not found: ${blockchainId}-${tokenAddress}`);
    }
    return decimals;
  }
}

describe('PackState', () => {
  test('pack_state_with_transitions', async () => {
    // Setup mock asset store
    const assetStore = new MockAssetStore();
    assetStore.addToken(42n, '0x90b7E285ab6cf4e3A2487669dba3E339dB8a3320' as Address, 8);
    assetStore.addToken(4242n, '0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2' as Address, 14);

    const channelId = '0x3e9dd25a843e3a234c278c6f3fab3983949e2404b276cacb3c47ada06e00f74b';

    const decimalFromString = (s: string): Decimal => {
      return new Decimal(s);
    };

    // Create a state with transitions (metadata will be derived from transitions)
    const state: State = {
      id: '',
      version: 24n,
      asset: 'test', // won't be used
      userWallet: '0x0' as Address,
      epoch: 0n,
      homeChannelId: channelId,
      transitions: [
        newTransition(TransitionType.HomeWithdrawal, 'tx1', 'account1', new Decimal(-100)),
        newTransition(TransitionType.EscrowWithdraw, 'tx2', 'account2', new Decimal(-50)),
      ],
      homeLedger: {
        blockchainId: 42n,
        tokenAddress: '0x90b7E285ab6cf4e3A2487669dba3E339dB8a3320' as Address,
        userBalance: decimalFromString('3'),
        userNetFlow: decimalFromString('2.00000001'),
        nodeBalance: decimalFromString('0'),
        nodeNetFlow: decimalFromString('-0.99999999'),
      },
      escrowLedger: {
        blockchainId: 4242n,
        tokenAddress: '0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2' as Address,
        userBalance: decimalFromString('3'),
        userNetFlow: decimalFromString('2.00000001'),
        nodeBalance: decimalFromString('0'),
        nodeNetFlow: decimalFromString('-0.99999999'),
      },
    };

    const packer = newStatePackerV1(assetStore);
    const packed = await packer.packState(state);
    expect(packed).toBeDefined();

    const expectedPackedState =
      '0x3e9dd25a843e3a234c278c6f3fab3983949e2404b276cacb3c47ada06e00f74b00000000000000000000000000000000000000000000000000000000000000180000000000000000000000000000000000000000000000000000000000000007fec8addc7e168fd4c7ae21f327ebd7981387be08f6b77786382887f56fd0e054000000000000000000000000000000000000000000000000000000000000002a00000000000000000000000090b7e285ab6cf4e3a2487669dba3e339db8a332000000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000011e1a300000000000000000000000000000000000000000000000000000000000bebc2010000000000000000000000000000000000000000000000000000000000000000fffffffffffffffffffffffffffffffffffffffffffffffffffffffffa0a1f010000000000000000000000000000000000000000000000000000000000001092000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000110d9316ec0000000000000000000000000000000000000000000000000000000b5e62103c2400000000000000000000000000000000000000000000000000000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffa50cef950240';

    expect(packed).toBe(expectedPackedState);
  });

   test('pack_state_without_escrow', async () => {
    // Setup mock asset store
    const assetStore = new MockAssetStore();
    assetStore.addToken(42n, '0x90b7E285ab6cf4e3A2487669dba3E339dB8a3320' as Address, 8);
    assetStore.addToken(4242n, '0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2' as Address, 14);

    const channelId = '0x3e9dd25a843e3a234c278c6f3fab3983949e2404b276cacb3c47ada06e00f74b';

    const decimalFromString = (s: string): Decimal => {
      return new Decimal(s);
    };

    // Create a state with transitions (metadata will be derived from transitions)
    const state: State = {
      id: '',
      version: 24n,
      asset: 'test',
      userWallet: '0x0' as Address,
      epoch: 0n,
      transitions: [
        newTransition(TransitionType.HomeDeposit, 'tx123', 'account456', new Decimal(1000)),
      ],
      homeChannelId: channelId,
      homeLedger: {
        blockchainId: 42n,
        tokenAddress: '0x90b7E285ab6cf4e3A2487669dba3E339dB8a3320' as Address,
        userBalance: decimalFromString('3'),
        userNetFlow: decimalFromString('2.00000001'),
        nodeBalance: decimalFromString('0'),
        nodeNetFlow: decimalFromString('-0.99999999'),
      },
      escrowLedger: undefined,
    };

    const packer = newStatePackerV1(assetStore);
    const packed = await packer.packState(state);
    expect(packed).toBeDefined();

    const expectedPackedState =
      '0x3e9dd25a843e3a234c278c6f3fab3983949e2404b276cacb3c47ada06e00f74b0000000000000000000000000000000000000000000000000000000000000018000000000000000000000000000000000000000000000000000000000000000234cf7c712e3949d1cd9bdf9403149be906146248f9401df1f393f12e6f209355000000000000000000000000000000000000000000000000000000000000002a00000000000000000000000090b7e285ab6cf4e3a2487669dba3e339db8a332000000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000011e1a300000000000000000000000000000000000000000000000000000000000bebc2010000000000000000000000000000000000000000000000000000000000000000fffffffffffffffffffffffffffffffffffffffffffffffffffffffffa0a1f010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000';

    expect(packed).toBe(expectedPackedState);
  });
});
