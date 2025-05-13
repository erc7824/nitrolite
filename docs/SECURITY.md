# This document outlines protocol-level security mechanisms

## Custody `IChannel` implementation

### Close

- HORIZONTAL: what is the last known on-chain state of a channel being closed
- VERTICAL: what is the status of a channel being closed
- CHANOPEN: a magic number showing a state is the first deposit state
- CHANCLOSE: a magic number showing a state is final

Here is an overview of possible scenarios a channel might be in while being closed:

|         | CHANOPEN not all joined | CHANOPEN all joined | operatable state | CHANCLOSE |
| ------- | ----------------------- | ------------------- | ---------------- | --------- |
| INITIAL | ❌                      | ❌                  | ❌               | ❌        |
| ACTIVE  | ❌                      | ✅                  | ✅               | ❌        |
| DISPUTE | ✅                      | ✅                  | ✅               | ❌        |
| FINAL   | ❌                      | ❌                  | ❌               | ❌        |

> NOTE: "not all joined" and "all joined" are characteristics of CHANOPEN state and can be determined by the number of signatures in the state.

> NOTE: it does not matter what state X is in, because as the Final state is signed by all participants, transition to it from the correct status is always valid.

As you can see, the channel can not be finalized when in an INITIAL or already FINAL status or while having CHANCLOSE state.

Let's review each scenario in detail.
X: an on-chain state the channel is in.
Y: a state the channel is being finalized with.

General rule is that Y must be CHANCLOSE.
General flow after all checks is:

- `meta.lastValidState = Y`
- `meta.status = FINAL`
- `unlock all allocations`
- `remove channel from participants legders`
- ...

#### INITIAL status

> The main goal: to verify Y is signed by all participants and is a valid CHANCLOSE state.

```md
- verify Y is a valid CHANCLOSE state (has CHANCLOSE magic number)
  verify Y has valid signatures from all participants
```

#### DISPUTE status

> The main goal: either close the channel after challenge or finalize the channel during challenge.

```md
- if if block.timestamp < meta.challengeExpire:
  - verify Y is a valid CHANCLOSE state (has CHANCLOSE magic number)
    verify Y has valid signatures from all participants
```

### Challenge

- HORIZONTAL: what is the last known on-chain state of a channel being challenged
- VERTICAL: what is the status of a channel being challenged
- CHANOPEN: a magic number showing a state is the first deposit state
- CHANCLOSE: a magic number showing a state is final

Here is an overview of possible scenarios a channel might be in while being challenged:

|         | CHANOPEN not all joined | CHANOPEN all joined | operatable state | CHANCLOSE |
| ------- | ----------------------- | ------------------- | ---------------- | --------- |
| INITIAL | ✅                      | ❌                  | ❌               | ❌        |
| ACTIVE  | ❌                      | ✅                  | ✅               | ❌        |
| DISPUTE | ✅                      | ✅                  | ✅               | ❌        |
| FINAL   | ❌                      | ❌                  | ❌               | ❌        |

> NOTE: "not all joined" and "all joined" are characteristics of CHANOPEN state and can be determined by the number of signatures in the state.

> NOTE: when influencing a channel with `checkpoint` or `challenge`, the current on-chain situation may have changed since the one pointed in by `meta.lastValidState`,
> i.e. some parties may have joined the channel.

As you can see, the channel can not be challenged in a FINAL status or while having CHANCLOSE state.

Let's review each scenario in detail.
X: an on-chain state the channel is in.
Y: a state the channel is being challenged with.

General rule is that Y can not be CHANCLOSE. In such case a user should call `close` function.
Another general rule is that after all checks there are the following on-chain changes:

- `meta.status = DISPUTE`
- `meta.lastValidState = Y`
- `meta.challengeExpire = block.timestamp + meta.chan.challenge`

#### INITIAL status

> The main goal: to verify Y is valid and >= X.

```md
- X is "CHANOPEN not all joined":
  - if (Y is CHANOPEN):
    - verify Y is a valid CHANOPEN state (has no proof, has CHANOPEN magic number)
      verify Y has not less signatures than X
      verify all Y signatures are valid
      verify all participants that supplied a signature in Y have deposited
  - else
    - verify all participants have deposited
      verify adjudicate(Y, proof)
```

#### ACTIVE status

> The main goal: to verify Y is valid and >= X.

```md
- X is "CHANOPEN all joined":
  - if (Y is CHANOPEN):
    - verify Y is a valid CHANOPEN state
      verify Y contains all signatures
      verify all Y signatures are valid
  - else
    - verify adjudicate(Y, proof)
- X is operatable state:
  - verify Y is NOT CHANOPEN
    verify NOT isMoreRecent(X, Y) (Y it not older than X)
    verify adjudicate(Y, proof)
```

#### DISPUTE status

> The main goal: to verify Y is valid and > X.

```md
- X is "CHANOPEN not all joined":
  - if (Y is CHANOPEN):
    - verify Y is a valid CHANOPEN state
      verify Y has more signatures than X
      verify all Y signatures are valid
      verify all participants that supplied a signature in Y have deposited
  - else
    - verify all participants have deposited
      verify adjudicate(Y, proof)
- X is "CHANOPEN all joined":
  - if (Y is CHANOPEN):
    - verify Y is a valid CHANOPEN state
      verify Y contains all signatures
      verify all Y signatures are valid
  - else
    - verify adjudicate(Y, proof)
- X is operatable state:
  - verify Y is NOT CHANOPEN
    verify isMoreRecent(Y, X)
    verify adjudicate(Y, proof)
```

### Checkpoint

- HORIZONTAL: what is the last known on-chain state of a channel being checkpointed
- VERTICAL: what is the status of a channel being checkpointed
- CHANOPEN: a magic number showing a state is the first deposit state
- CHANCLOSE: a magic number showing a state is final

Here is an overview of possible scenarios a channel might be in while being checkpointed:

|         | CHANOPEN not all joined | CHANOPEN all joined | operatable state | CHANCLOSE |
| ------- | ----------------------- | ------------------- | ---------------- | --------- |
| INITIAL | ✅                      | ❌                  | ❌               | ❌        |
| ACTIVE  | ❌                      | ✅                  | ✅               | ❌        |
| DISPUTE | ✅                      | ✅                  | ✅               | ❌        |
| FINAL   | ❌                      | ❌                  | ❌               | ❌        |

> NOTE: "not all joined" and "all joined" are characteristics of CHANOPEN state and can be determined by the number of signatures in the state.

> NOTE: when influencing a channel with `checkpoint` or `challenge`, the current on-chain situation may have changed since the one pointed in by `meta.lastValidState`,
> i.e. some parties may have joined the channel.

As you can see, the channel can not be checkpointed in a FINAL status or while having CHANCLOSE state.

Let's review each scenario in detail.
X: an on-chain state the channel is in.
Y: a state the channel is being challenged with.

General checks:

- Y is not CHANCLOSE. In such case a user should call `close` function.
- Y is not CHANOPEN. In such case a user should call `join` function.

Another general rule is that after all checks there are the following on-chain changes:

- `meta.status = updatedStatus`, where the latter is determined during checks.
- `meta.lastValidState = Y`

> The main goal: to verify Y is valid and > X.

#### INITIAL status

```md
- X is "CHANOPEN not all joined":
  - verify all participants have deposited
    verify adjudicate(Y, proof)

updatedStatus = ACTIVE
```

#### ACTIVE status

```md
- X is "CHANOPEN all joined":
  - verify adjudicate(Y, proof)
- X is operatable state:
  - verify isMoreRecent(Y, X)
    verify adjudicate(Y, proof)

updatedStatus = ACTIVE
```

#### DISPUTE status

```md
- X is "CHANOPEN not all joined":
  - verify all participants have deposited
    verify adjudicate(Y, proof)
- X is "CHANOPEN all joined":
  - verify adjudicate(Y, proof)
- X is operatable state:
  - verify isMoreRecent(Y, X)
    verify adjudicate(Y, proof)

updatedStatus = ACTIVE
meta.challengeExpire = 0
```
