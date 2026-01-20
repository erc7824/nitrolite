# Security properties of on-chain Nitrolite protocol infrastructure

## Behavior

These are behavior rules of the Clearnode or the logic how a user (should) operate.

1. if challenged with an older state, then checkpoint with the latest one

This produces the following invariant:
> A channel can only be challenged with the latest known (even off-chain) state.

---

2. if Node is low on liquidity (below some threshold), User checkpoints latest off-chain state, and optionally closes the channel
(Or User requests to migrate channel to another chain where Node has liquidity)

Invariant:
> The Node always have funds to transfer to the User IN-BETWEEN OPERATIONS
(this is NOT TRUE for non-home chain deposit, -//- withdrawal or a home chain migration, please see below).

---

3. Node stops issuing new states when NON-HOME chain deposit, -//- withdrawal or a home chain migration has started and not yet finished

---

4. Both `cross-chain withdrawal` and `home-chain` migration end with a state pushed to a non-home chain, and
  `cross-chain deposit` results in either funds automatically unlocked for Node, or an already signed state that an unlock them.

Given the 3 and 4, an invariant:
> at any moment of time a CCTB state will certainly contain not more than 2 per-chain states.

---

5. A party never signs a state with a `version` that was already signed for this channel.

Invariant:
> No different states with the same `version` can exist for the same channel.

## Invariants

---

- (NOT TRUE) only less-or-equal amount of internally-accounted funds can be withdrawn (NOT TRUE for states that include "receive" off-chain ops)

The absence of the aforementioned invariant creates a huge risk of an attacker draining the Node.
To protect from this, the Node should keep CORRECT track of off-chain user funds.
CAUTION IS REQUIRED.

P.S. This invariant still can be enforced by updating `lockedFunds` per channel meta-variable during on-chain state processing,
e.g. when processing "receive X, withdraw Y", increase `lockedFunds` (and "lock" Node's funds in channel) by X, and then decrease by Y.

---

- User funds can be withdrawn only after channel is finalized (closed or challenged) or during WITHDRAW action
- any action is valid only with a Node's signature (for now, but this condition may be loosened to improve UX by making protocol more complex)
- a state with `version` <= `latestKnownVersion` per chain cannot be accepted as valid
- for challenge a state with `version` < `latestKnownVersion` per chain cannot be accepted as valid
- a channel with the same `channelId` cannot be created twice
- an escrow with the same `escrowId` cannot be created twice
- on-chain-stored state has already been processed

---

## Formal Invariants List

### Channel identity and authorization

1. **Channel uniqueness**: A channel identified by `channelId = hash(Definition)` can be created at most once.
2. **Signature authorization**: Every enforceable state must be signed by both User and Node (unless explicitly relaxed in future versions).
3. **Version monotonicity**: For a given channel, every valid state has a strictly increasing `version`.
4. **Version uniqueness**: No two different states with the same `version` may exist for the same channel.

---

### State validity

5. **Per-chain correctness**: For any per-chain state, allocations and net flows are internally consistent and non-negative where required by the chain role (home vs non-home).
6. **Single-chain enforcement (current scope)**: For single-chain operation, the home-state `chainId` must equal `block.chainid`.
7. **Allocation backing**: The sum of allocations in an enforced state must equal the amount of locked collateral implied by previous state plus net flow deltas.
8. **No retrogression**: A state with `version ≤ lastEnforcedVersion` cannot be enforced or checkpointed.

---

### Liquidity and accounting

9. **Locked funds safety**: Channel locked funds are never negative.
10. **Node liquidity constraint**: Whenever a state requires locking Node funds, the Node must have sufficient available on-chain liquidity at enforcement time.
11. **Controlled imbalance**: User or Node net flows may temporarily exceed allocations only during explicitly allowed escrow or migration phases.

---

### Operational semantics

12. **Deposit semantics**: A state with intent `DEPOSIT` must include a positive user net-flow delta.
13. **Withdrawal semantics**: A state with intent `WITHDRAW` must include a negative user net-flow delta and must not increase user allocation beyond previous allocation.
14. **Operate / checkpoint semantics**: A state with intent `OPERATE` must not change user net flow on the enforcing chain.
15. **Close semantics**: A state with intent `CLOSE` finalizes the channel and distributes allocations to both parties.

---

### Challenge mechanism

16. **Challenge admissibility**: A channel can only be challenged when in `OPERATING` state.
17. **Latest-state challenge rule**: A challenge must reference a state with `version ≥ lastEnforcedVersion`; if higher, that state is enforced first.
18. **Challenge resolution**: Any strictly newer valid state supersedes an active challenge and returns the channel to `OPERATING`.
19. **Challenge finality**: If no newer state is enforced before challenge expiry, the channel may be unilaterally closed using the last enforced state.

---

### Cross-chain and multi-state structure

20. **Bounded per-chain states**: At any moment, a cross-chain channel state contains at most two per-chain states (home and non-home).
21. **Flow suspension**: During escrow deposit, escrow withdrawal, or migration, the Node must not issue new states until completion or challenge.
22. **Recoverability**: Every escrow or migration phase must be completable or revertible via timeout and challenge on at least one chain.

---

### Safety guarantees

23. **Enforcement determinism**: Enforcing the same `(prevState, candidateState)` pair always yields the same on-chain result.
24. **Invariant preservation**: Every state transition that can be enforced on-chain preserves all invariants listed above.
25. **Latest-state dominance**: The economically correct outcome is always determined by the latest valid signed state, regardless of enforcement order.
