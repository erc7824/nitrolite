# Application Sessions

Previous: [Extensions Overview](overview.md)

---

This document describes the application session extension supported by the Nitrolite protocol.

## Purpose

Application sessions enable off-chain application functionality within the protocol. They allow participants to run applications that manage committed assets according to application-specific rules.

Examples include games, exchanges, and other interactive applications that require fast off-chain state updates.

## Application Session Entity

An application session is defined by:

```
AppSession {
  SessionId:      bytes32         // unique session identifier
  ChannelId:      bytes32         // parent channel identifier
  Participants:   []address       // session participants
  Application:    address         // application logic identifier
  CommittedAssets: []Allocation   // assets committed from the channel
}
```

The session identifier is derived deterministically from the session parameters.

## Application State

Each application session maintains its own state:

```
AppState {
  SessionId:    bytes32       // parent session identifier
  Version:      uint64        // state version
  Allocations:  []Allocation  // current asset distribution within the session
  Data:         bytes         // application-specific data
}
```

Application state follows the same versioning rules as channel state:

- Versions are strictly increasing
- Each update requires valid signatures
- The latest signed state is enforceable

## Application Session Keys

Participants may delegate signing authority for application session operations to session keys.

Session key authorization:

1. The participant's primary key signs an authorization granting the session key signing rights for a specific session
2. The session key may then sign application state updates on behalf of the participant
3. Session key authorization is scoped to a specific session and may include expiration

Session keys enable applications to sign state updates without requiring the participant's primary key for each operation.

## Commit Operation

The commit operation moves assets from a channel into an application session.

Process:

1. Participants agree on session parameters and the amount to commit
2. A channel state update is created with a commit transition, reducing channel allocations
3. An application session is created with the committed assets
4. Both the channel state update and session creation are signed atomically

Rules:

- The committed amount must not exceed the participant's channel allocation
- All channel participants must sign the commit transition
- The application session must reference a valid application identifier

## Release Operation

The release operation returns assets from an application session back to the channel.

Process:

1. The application session reaches a terminal state (application completed or participants agree to close)
2. A release transition is created, specifying the final asset distribution
3. The channel state is updated to reflect the returned assets
4. Both the release and channel state update are signed atomically

Rules:

- Released assets must not exceed the total committed to the session
- The release distribution must be authorized by the application state
- All channel participants must sign the release transition

## Interaction with Channel Protocol

Application sessions coordinate with the channel protocol as follows:

- **Asset consistency** — the sum of channel allocations and all active session commitments must equal the total channel deposits
- **State ordering** — commit and release transitions follow standard channel state versioning rules
- **Enforcement** — if a channel is enforced on-chain, active application sessions may be resolved according to their latest signed state
- **Independence** — multiple application sessions may be active simultaneously on a single channel

## Current Limitations

The current implementation has the following limitations:

- Application logic is not enforced on-chain; resolution depends on the latest mutually signed application state
- Session keys cannot be revoked before their expiration without a channel state update
- The number of concurrent application sessions per channel may be limited
- Application-specific dispute resolution is not yet supported on the settlement layer

---

Previous: [Extensions Overview](overview.md)
