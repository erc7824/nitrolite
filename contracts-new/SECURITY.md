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
(this it NOT TRUE for non-home chain deposit, -//- withdrawal or a home chain migration, please see below).

---

3. Node stops issuing new states when NON-HOME chain deposit, -//- withdrawal or a home chain migration has started and not yet finished

---

4. Both `cross-chain withdrawal` and `home-chain` migration end with a state pushed to a non-home chain, and
  `cross-chain deposit` results in either funds automatically unlocked for Node, or an already signed state that an unlock them.

Given the 3 and 4, an invariant:
> at any moment of time a CCTB state will certainly contain not more than 2 per-chain states.

---

## Invariants

---

- (NOT TRUE) only less-or-equal amount of internally-accounted funds can be withdrawn (NOT TRUE for states that include "receive" off-chain ops)

The absense of the beforementioned invariant creates a huge risk of an attacker draining the Node.
To protect from this, the Node should keep CORRECT track of off-chain user funds.
CAUTION IS REQUIRED.

P.S. This invariant still can be enforced by updating `lockedFunds` per channel meta-variable during on-chain state processing,
e.g. when processing "receive X, withdraw Y", increase `lockedFunds` (and "lock" Node's funds in channel) by X, and then decrease by Y.

---

- User funds can be withdrawn only after channel is finalized (closed or challenged) or during WITHDRAW action
- any action is valid only with a Node's signature (for now, but this condition may be loosen to improve UX by making protocol more complex)
- a state with `version` <= `latestKnownVersion` per chain can not be accepted as valid
- for challenge a state with `version` < `latestKnownVersion` per chain can not be accepted as valid
- a channel with the same `channelId` can not be created twice
- an escrow with the same `escrowId` can not be created twice

---

It is easy to implement the protocol without NON-home chain operations, as all on-chain operations will
end up submitting a state on-chain, while off-chain ops (transfer, app-session ops) will just produce an off-chain
state with a changed amount.
