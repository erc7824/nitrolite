# Cross-Chain and Asset Model

Previous: [Enforcement and Settlement](enforcement-and-settlement.md) | Next: [Interactions](interactions.md)

---

This document describes the unified asset model and cross-chain functionality.

## Purpose

The unified asset model allows participants to operate on assets from multiple blockchains within a single channel. This eliminates the need for separate channels per blockchain and enables cross-chain interactions.

## Unified Asset Concept

Assets in the Nitrolite protocol are identified independently of any specific blockchain.

A unified asset is defined by:

```
Asset {
  Symbol:   string    // canonical asset identifier (e.g., "USDC")
  Decimals: uint8     // decimal precision
}
```

The same logical asset may exist on multiple blockchains. The protocol treats all instances of a unified asset as fungible within a channel.

## Home Chain

Every channel has a designated home chain.

The home chain is:

- the blockchain where the channel was created
- the authoritative source for state enforcement
- the chain where the settlement contract exists

The home chain is fixed at channel creation and determines where enforcement operations are executed.

## Home and Non-Home Ledger Roles

**Home Ledger**
The home ledger is the primary record of asset allocations. It is associated with the home chain and is directly enforceable through the settlement contract.

Responsibilities:

- tracks the authoritative asset allocations
- receives checkpoints for enforcement
- holds deposited assets in the settlement contract

**Non-Home Ledger**
A non-home ledger tracks asset allocations on a blockchain other than the home chain.

Responsibilities:

- tracks assets deposited from non-home chains
- reflects cross-chain deposit and withdrawal operations
- coordinates with the home ledger for consistency

## Cross-Chain Deposit Rules

To deposit assets from a non-home chain into a channel:

1. The participant deposits assets into the settlement contract on the non-home chain
2. The non-home ledger records the deposit
3. A corresponding state update reflects the new allocation in the channel state
4. The home ledger is updated to include the cross-chain deposit in the unified view

## Cross-Chain Withdrawal Rules

To withdraw assets to a non-home chain:

1. The channel state is updated to reduce the participant's allocation
2. The non-home ledger records the withdrawal
3. The settlement contract on the non-home chain releases the assets to the participant

## Ledger Swap Rules

Assets may be transferred between ledgers within a channel.

Rules:

- The total amount of a unified asset across all ledgers must remain constant
- Ledger swaps require signatures from all participants
- The swap is reflected in both the source and destination ledger allocations

## Home Chain Migration

The home chain of a channel may be changed under the following conditions:

1. All participants agree to the migration
2. The channel state is checkpointed and finalized on the current home chain
3. A new channel is created on the target chain with equivalent state
4. Assets are transferred from the old home chain to the new home chain

Home chain migration is an expensive operation and is not expected during normal operation.

## Cross-Chain Replay Protection

The protocol prevents cross-chain replay through:

- **Chain Identifier** — each ledger component is bound to a specific chain identifier
- **Channel Identifier** — channel identifiers include home chain information
- **Settlement Contract Address** — enforcement operations target a specific contract on a specific chain

## Current Version Notes

In the current protocol version:

- Cross-chain operations require trust in the node to relay state correctly between chains
- Full cross-chain enforcement (trustless bridging) is a planned future improvement
- The number of supported non-home chains may be limited by implementation

---

Previous: [Enforcement and Settlement](enforcement-and-settlement.md) | Next: [Interactions](interactions.md)
