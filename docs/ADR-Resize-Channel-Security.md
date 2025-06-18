# Architectural Decision Record: Channel Resize Security

## Status
Proposed

## Date
2025-03-06

## Context
The Nitrolite system supports payment channels across multiple blockchain networks through the Clearnode server. Users can resize their channels to adjust the allocated funds. Currently, the resize flow works as follows:

1. User calls the Clearnode API with a resize request
2. Clearnode validates the request and user balance
3. Clearnode generates and signs a new state with RESIZE intent
4. Clearnode returns the signed state to the user
5. User signs the state and submits it to the smart contract

This process has revealed a vulnerability: users can double-spend their funds by requesting resizes across multiple chains that collectively exceed their total balance. Since the server signs states that remain valid indefinitely and doesn't track pending state commitments across chains, a user could:

1. Deposit 5 USDC on Chain 1 and 5 USDC on Chain 2 (total balance of 10 USDC)
2. Request a resize of Channel 1 from 5 to 10 USDC (server signs this state)
3. Request a resize of Channel 2 from 5 to 10 USDC (server signs this state too)
4. Submit both states to their respective blockchains
5. Result: User has committed 20 USDC despite only having 10 USDC in their balance

## Decision Drivers
* Security: Prevent double-spending across multiple chains
* User Experience: Maintain a seamless resize process
* System Integrity: Ensure balance accounting remains accurate
* Technical Feasibility: Implement with minimal changes to existing contracts
* Operational Overhead: Consider the impact on server operations

## Considered Options

### Option 1: Timeout-Based Fund Locking
* Lock funds in the database when a resize is requested
* Set a timeout to release locks if not confirmed on-chain
* Challenge: The smart contract has no notion of state expiration, so users could still submit expired states

### Option 2: State Versioning with Registry
* Maintain a server-side registry of valid state versions
* When signing a new state, include a global version number in state data
* When processing on-chain events, invalidate all previous versions
* Challenge: **Critical flaw** - Once a state is signed by the server, it can be submitted to the blockchain regardless of database state. Without smart contract modifications to validate against the external registry, this approach only detects double-spending after it occurs but doesn't prevent it.

### Option 3: Cross-Chain Lock Commitments
* Modify smart contracts to require a "lock" transaction on a primary chain
* This lock creates a verifiable on-chain commitment of total funds allocation
* Other chains verify against this commitment before processing resizes
* Each resize reduces the available commitment amount
* Challenge: Requires smart contract modifications, adds cross-chain dependencies, and significantly complicates the user experience with additional transactions

### Option 4: Server-Side Blockchain Submission
* Instead of returning signed states to users, server directly submits resize transactions
* Server tracks on-chain confirmations before allowing additional resizes
* Server manages transaction queuing, gas prices, and retries

## Decision
**We want to implement Option 4: Server-Side Blockchain Submission for channel resize operations.**

This approach offers the strongest security guarantees while maintaining a good user experience. The server will directly submit resize transactions to the blockchain after validating user requests, eliminating the window for double-spending.

After thorough analysis, we determined that:
- Option 1 (Timeout-Based Fund Locking) fails because signed states remain valid indefinitely
- Option 2 (State Versioning) cannot prevent double-spending without smart contract modifications
- Option 3 (Cross-Chain Lock Commitments) would work but requires significant contract changes and degrades UX

Server-side submission (Option 4) addresses the vulnerability without contract modifications by ensuring users never receive multiple valid states they could submit independently.

## Implementation Details
1. Modify the `HandleResizeChannel` function to:
   - Accept user's signed resize request
   - Validate the request and check balance
   - Sign the state as before
   - Submit the transaction directly to the blockchain
   - Return transaction hash to the user instead of the signed state

2. Add transaction tracking:
   - Create a `pending_transactions` table to track submitted transactions
   - Update channel status only after on-chain confirmation
   - Implement retry mechanism for failed transactions

3. Balance management:
   - Lock funds immediately when resize is initiated
   - Release locks only if transaction permanently fails
   - Track transaction status through event monitoring

## Consequences

### Positive
* Eliminates the cross-chain double-spend vulnerability
* Simplifies the user experience (no need to handle blockchain transactions)
* Provides stronger consistency between database and blockchain state
* Allows for better transaction management (gas optimization, retries)

### Negative
* Requires secure management of server signing keys
* Adds complexity to server-side transaction management
* Creates dependency on server for all resize operations
* Requires implementation of transaction fee charging mechanism

### Neutral
* Shifts responsibility for transaction submission from users to the service provider
* Transaction costs can be charged to user balances
* May require adjustments to the service fee structure to account for gas costs and service overhead

## Open Questions
1. What specific mechanism will be used to charge transaction fees to user balances?
   - Fixed fee per resize
   - Dynamic fee based on actual gas costs
   - Fee plus margin model
   
2. How to handle transaction failures?
   - Maximum number of retry attempts
   - Escalation process for persistently failing transactions
   - Refund policy for failed transactions

3. What happens if the server goes down during transaction processing?
   - Recovery mechanism for in-flight transactions
   - Consistency checks after system recovery

4. How will we handle network congestion and transaction pricing?
   - Gas price strategy during high network congestion
   - Maximum gas price thresholds
   - Prioritization of transactions

5. What monitoring and alerting should be implemented for transaction tracking?
   - Real-time monitoring of pending transactions
   - Alerts for stuck or failed transactions
   - Periodic reconciliation between database and blockchain state
