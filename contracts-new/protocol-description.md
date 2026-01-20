# Nitrolite Protocol — On-Chain and Off-Chain Architecture

## High-level goal

Nitrolite is an **extended state-channel protocol** that enables:

* continuous off-chain transfers and application interactions,
* frequent on-chain settlement (deposit / withdrawal),
* cross-chain liquidity movement (bridging),
* without locking all funds until channel closure.

The protocol trades atomic cross-chain guarantees for **optimistic enforcement with challenge recovery**, relying on cryptographic authorization and game-theoretic incentives.

---

## Core abstraction: Cross-Chain Token Balance (CCTB) State

A **channel** between a **User** and a **Node** is represented by a monotonically increasing sequence of **Cross-Chain States**.

Each state:

* has a strictly increasing `version`,
* is signed by both User and Node,
* encodes the net result of:

  * on-chain operations (deposit / withdrawal / migration),
  * off-chain transfers,
  * off-chain application sessions,
  * escrow preparation and execution.

A state may refer to **multiple chains**, but at any time only **at most two per-chain sub-states** exist (home and non-home).

Each **per-chain sub-state** represents accounting on a specific chain and consists of:

* **absolute allocations**
  (`userAllocation`, `nodeAllocation`) that must be fully backed by collateral locked on that chain, and

* **cumulative net flows**
  (`userNetFlow`, `nodeNetFlow`) that encode the aggregate effect of deposits, withdrawals, off-chain transfers, and app-session lock/unlock events since channel creation.

The difference between successive states’ net flows determines how much value must be pulled from or pushed to each party during on-chain enforcement.

---

## Off-chain protocol (control plane)

### Participants

* **User** — owns funds and initiates actions.
* **Node (Broker)** — provides liquidity, routing, and coordination.

### Off-chain responsibilities

The off-chain protocol is responsible for:

1. **State construction**

   * The Node aggregates:

     * off-chain transfers,
     * app-session lock/unlock events,
     * pending on-chain actions.
   * These are netted into a new `CrossChainState` by updating per-chain allocations and cumulative net flows.

2. **State authorization**

   * Both User and Node sign the full state:

     ```
     (channelId, version, intent, homeState, nonHomeState)
     ```
   * A party **never signs two different states with the same version**.

3. **Liquidity enforcement (Node responsibility)**

   * The Node must ensure it has enough liquidity to back absolute allocations:

     * between normal operations,
     * except during explicitly allowed escrow or migration phases.
   * If liquidity drops below a threshold, the User may:

     * checkpoint the latest state on-chain,
     * withdraw,
     * or migrate the channel.

4. **Flow control**

   * When a cross-chain escrow or migration is in progress:

     * the Node **stops issuing new states**,
     * until the process completes or is challenged.

5. **Optimistic bridging**

   * Cross-chain actions are **not atomically verifiable** on-chain.
   * Correctness is ensured by:

     * signed states,
     * cumulative net-flow accounting,
     * timeouts,
     * challenge rights.

---

## Off-chain actions encoded in states

### Off-chain transfers

* When a User **sends** funds off-chain:

  * user allocation decreases,
  * node net flow increases.
* When a User **receives** funds off-chain:

  * user allocation increases,
  * node net flow decreases.

These changes are reflected only in cumulative net flows until enforced on-chain.

---

### Off-chain application sessions

* App sessions are off-chain sub-channels governed by an external server.
* Funds may be:

  * **locked** into a session (flow to Node),
  * **unlocked** from a session (flow to User).
* Only signatures are required for persistence.
* Session effects are netted into cumulative net flows of the next enforceable state.

---

## On-chain protocol (enforcement plane)

The on-chain contract is the **final arbiter** of correctness.

It does not reconstruct intent — it **verifies and enforces signed states** by:

* validating signatures and monotonic versioning,
* applying the delta between the last enforced state and the submitted state,
* pulling or pushing funds according to net-flow differences,
* updating locked collateral to match absolute allocations.

---

## Channel lifecycle (on-chain)

### 1. Channel creation

* A channel is created with an initial signed state:

  * version = 0,
  * intent = CREATE,
  * funds pulled from the User (home chain).
* Channel enters `OPERATING`.

---

### 2. Normal operation (OPERATING)

While operating:

* Any **newer signed state** may be enforced on-chain.
* Enforcement may:

  * pull funds from User,
  * push funds to User,
  * lock or unlock Node liquidity.
* Enforcement may occur for:

  * deposit,
  * withdrawal,
  * checkpoint,
  * escrow execution,
  * migration execution.

Off-chain activity can continue indefinitely between enforcements.

---

### 3. Deposit (single-chain)

* User signs a state with intent = DEPOSIT.
* User net flow becomes positive.
* On enforcement:

  * funds are pulled from User,
  * locked into the channel.

---

### 4. Withdrawal (single-chain)

* User signs a state with intent = WITHDRAW.
* User net flow becomes negative.
* On enforcement:

  * funds are pushed to User,
  * channel locked funds decrease.

---

### 5. Checkpoint

* A state with intent = OPERATE.
* User net flow delta must be zero.
* Used to:

  * acknowledge off-chain transfers,
  * clear challenges,
  * synchronize cumulative net-flow accounting.

---

## Challenge mechanism (optimistic safety)

### Purpose

Challenges protect against:

* submission of outdated states,
* malicious or crashed counterparties,
* incomplete cross-chain operations.

---

### Challenge rules

* Only channels in `OPERATING` can be challenged.
* A challenge references a signed state.
* If the challenged state is **older than the latest signed state**:

  * the newest valid signed state **must be enforced first**, regardless of its intent.

Invariant:

> Dispute resolution always requires processing the newest valid signed state, even if that state represents escrow execution or migration rather than deposit or withdrawal.

---

### Resolving a challenge

* Any party may submit a **strictly newer signed state**.
* If valid:

  * it is enforced,
  * net-flow deltas are applied,
  * the challenge is cleared,
  * channel returns to `OPERATING`.

---

### Challenge timeout

* If no newer state is submitted before expiry:

  * channel may be closed unilaterally,
  * allocations are paid out according to the last enforced state.

---

## Channel closure

A channel can be closed:

1. **Cooperatively**

   * via a signed CLOSE state.

2. **Unilaterally**

   * after a challenge expires.

Closure:

* pushes all remaining allocations to User and Node,
* sets channel status to CLOSED.

---

## Cross-chain operations (bridging)

Cross-chain actions are **two-phase** and **optimistic**.

### Why two-phase?

Because:

* one chain cannot directly observe or verify another chain’s state,
* atomic enforcement is impossible without foreign-chain verification.

The protocol deliberately does **not** rely on light clients (on-chain verification of foreign headers, proofs, and validator signatures), as they are complex, expensive, and chain-specific.

The two phases are:

1. **Preparation phase**

   * liquidity is locked on chains where needed,
   * an escrow object (possible with timeouts) is created,
   * Node stops issuing new states.

2. **Execution phase**

    * an execution state that updates allocations and net flows is issued and signed
    * this state may be enforced immediately or later, but is enforceable to resolve disputes.

---

## Escrow deposit (bridging in)

### Preparation phase

* User locks funds on the **non-home chain**.
* Node locks equal liquidity on the **home chain**.
* An escrow object with timeouts is created.

---

### Execution phase

* A signed execution state updates allocations and net flows:

  * User’s non-home allocation decreases,
  * Node’s home allocation decreases,
  * corresponding net flows encode the swap.

This execution state **may be enforced immediately or later**, but must be enforceable to resolve disputes.

---

## Escrow withdrawal (bridging out)

### Preparation phase

* Node locks withdrawal liquidity on the **non-home chain**.

---

### Execution phase

* Signed state updates allocations and net flows so that:

  * User receives funds on the non-home chain.

If enforcement stalls:

* challenges and timeouts guarantee completion or reversion.

---

## Home chain migration

Migration is a special case of escrow withdrawal:

* User changes which chain is the “home” security chain.
* Node locks liquidity on the target chain.
* Final execution state:

  * releases Node liquidity on the old home chain,
  * establishes allocations and net flows on the new home chain.

This state may need to be enforced to clear challenges.

---

## Security model summary

* **Authorization**: all state changes require valid signatures.
* **Monotonicity**: `version` strictly increases.
* **Replay resistance**: no two states with the same version can coexist.
* **Liquidity safety**: absolute allocations must be collateral-backed.
* **Optimistic safety**:

  * challenges always resolve by enforcing the newest valid state,
  * stalled cross-chain operations can always be completed or reverted.

---

## Mental model

* Off-chain protocol **decides what should happen**.
* On-chain contract **enforces the latest authorized accounting state**.
* Bridging is **non-atomic but recoverable**.
* The channel is **continuously enforceable**, not locked until closure.
