# Clearnode: Key Entities

This document outlines the primary data entities in the Clearnode protocol. Understanding these entities and their relationships is essential for developers working with Clearnode.

## Channel

A Channel represents a state channel between participants for token transfers, enabling off-chain transactions with on-chain settlement capabilities.

**Fields:**
- `ChannelID` (string): Unique identifier for the channel
- `ChainID` (uint32): Blockchain network identifier
- `Token` (string): Token address used in this channel
- `Participant` (string): Address of the participant
- `Amount` (uint64): Current amount in the channel
- `Status` (enum): Current state of the channel ("joining", "open", "closed")
- `Challenge` (uint64): Challenge period for disputes (in blocks)
- `Nonce` (uint64): Sequence number for state updates
- `Version` (uint64): Version number for tracking protocol changes
- `Adjudicator` (string): Address of the adjudicator contract

Channels are uniquely identified by their ChannelID and are associated with specific Assets via Token and ChainID.

## Asset

An Asset represents a cryptocurrency or token that can be used in payment channels.

**Fields:**
- `Token` (string): Contract address of the token
- `ChainID` (uint32): Blockchain network identifier
- `Symbol` (string): Token symbol (e.g., "USDC")
- `Decimals` (uint8): Number of decimal places for the token

Assets are uniquely identified by the combination of their Token address and ChainID.

## AppSession

An AppSession represents a virtual payment application session between multiple participants, enabling complex payment applications beyond simple transfers.

**Fields:**
- `SessionID` (string): Unique identifier for the session
- `Protocol` (string): Protocol version used
- `Challenge` (uint64): Challenge period for disputes
- `Nonce` (uint64): Sequence number for state updates
- `Participants` (string[]): List of participant addresses
- `Weights` (int64[]): Voting weights of participants
- `Quorum` (uint64): Required consensus threshold
- `Version` (uint64): Version number
- `Status` (enum): Current state of the session (matches Channel status options)

AppSessions enable multi-party payment applications with consensus mechanisms through weighted signatures.

## Ledger Entry

A Ledger Entry records credits and debits for accounts, providing a complete audit trail of all financial operations.

**Fields:**
- `AccountID` (string): Identifier of the account
- `AccountType` (enum): Type of the account
- `AssetSymbol` (string): Symbol of the asset being tracked
- `Participant` (string): Address of the participant
- `Credit` (decimal): Amount credited
- `Debit` (decimal): Amount debited
- `CreatedAt` (timestamp): Creation timestamp

The ledger system maintains balances by tracking all credits and debits for each account-asset pair. Importantly, all ledger operations use decimal values to maintain precision, while blockchain-related operations (in Channels and other on-chain entities) use big.Int to ensure consistency with different tokens on different networks, each with their own decimal precision requirements.

## RPCRecord

An RPCRecord stores the history of RPC messages for auditing and retrieval.

**Fields:**
- `Sender` (string): Address of the sender
- `RequestID` (uint64): Request identifier
- `Method` (string): RPC method name
- `Params` (bytes): Serialized request parameters
- `Timestamp` (uint64): Unix timestamp
- `RequestSignature` (string[]): Request signatures
- `Response` (bytes): Serialized response
- `ResponseSignature` (string[]): Response signatures

RPCRecords provide a complete history of all protocol communications.

## NetworkConfig

A NetworkConfig represents configuration for a blockchain network.

**Fields:**
- `Name` (string): Network name (e.g., "mainnet", "goerli")
- `ChainID` (uint32): Blockchain network identifier
- `RpcURL` (string): RPC endpoint URL
- `CustodyAddress` (string): Address of the custody contract

NetworkConfig enables the protocol to interact with different blockchain networks.

## Entity Relationships

- **Channels** reference **Assets** via Token and ChainID.
- **AppSessions** are associated with multiple **Channels** through `participant` address.
- **Ledger Entries** track balances for participants in unified accounts and **AppSessions**.
- **RPCRecords** store the history of RPCMessages.
- **Custody** listens to blockchain events and updates **Channels** and **Ledger Entries** working with **Assets** to maintain token precision.

## Data Type Conventions

- **Blockchain Operations**: All values related to blockchain operations (Channel amounts, Channel Resize, Channel Close, on-chain transfers) use big.Int to maintain consistency with different tokens across networks, each with their own decimal precision requirements.
- **Ledger Operations**: All internal ledger operations use decimal values to maintain precision for financial accounting purposes.
- **Conversion**: When moving between on-chain and off-chain representations, values are converted using the Asset's `Decimals` field as the scaling factor.