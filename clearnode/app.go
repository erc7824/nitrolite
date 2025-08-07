package main

// Entities:
//   - User:     controls keys, global balances (Available/Reserved), withdraw nonce.
//   - App:      created/owned by a User.
//   - AppUser:  centralized per-App user profile/state (one row per (App,User)).
//   - Room:     App-defined shard/topic (formerly "Channel"); holds many RoomUsers.
//   - RoomUser: per-Room per-User row (data + in-room balances).
//
// Global invariants (per User, Asset):
//   Reserved == sum over all (App,Room) of RoomUser.Balances[asset]
//   Available >= 0
//   Room.Totals[asset] == sum over users of RoomUser.Balances[asset]
//
// Concurrency model:
//   - Row-local updates (RoomUser, AppUser) DO NOT carry PrevVersion; they only
//     bump the row's Nonce and must KEEP per-room totals constant.
//   - Structural updates (Join/Quit/Policy changes) REQUIRE PrevVersion on the
//     records they structurally modify; they bump Version after success.
//
// On-chain:
//   - Withdraw proofs reference only BalancesRoot (membership of (User,Asset,Available,WithdrawNonce)).
//   - Rooms/AppUser live off-chain; export = raw data + hash + membership proofs.

import (
	"github.com/shopspring/decimal"
)

// -----------------------------------------------------------------------------
// Identifiers & primitives
// -----------------------------------------------------------------------------

type (
	UserID [32]byte
	AppID  [32]byte
	RoomID [32]byte

	// AssetID identifies an asset inside the network.
	// Use a registry to map external chain/token to this compact ID.
	AssetID uint32
)

// Amount is a fixed-point decimal type for asset amounts.
type Amount = decimal.Decimal

// Signature is a generic wrapper. In production you’ll likely carry signer type
// (user/app), algorithm, and the public key or its ID.
type Signature struct {
	Alg   string // e.g. "secp256k1", "ed25519"
	Pub   []byte // optional; or store a key ID
	Sig   []byte
	Nonce uint64 // optional anti-replay at the transport layer
}

// -----------------------------------------------------------------------------
// Global balances (authority for withdrawals)
// -----------------------------------------------------------------------------

// GlobalBalance is the per-(User,Asset) record committed under BalancesRoot.
//   - Available: funds the user can transfer/withdraw out of rooms.
//   - Reserved:  funds locked inside rooms across ALL apps/rooms.
//   - Version:   CAS field for structural ops (Join/Quit/settlement) touching this row.
//   - WithdrawNonce: monotonically increments on successful on-chain withdrawals
//     to prevent replay of withdrawal proofs on target chains.
type GlobalBalance struct {
	Available     Amount
	Reserved      Amount
	Version       uint64
	WithdrawNonce uint64
}

// -----------------------------------------------------------------------------
// App & AppUser (centralized per-App user state)
// -----------------------------------------------------------------------------

// App is created/owned by a User. You can extend with metadata/policies.
type App struct {
	ID        AppID
	Owner     UserID // who controls room creation, policies, treasury, etc.
	CreatedAt int64  // unix seconds
}

// AppUser is the centralized, per-App state for a given user.
// - Nonce: increments on any update to this row (anti-replay).
// - Data: arbitrary app-scoped data (e.g., user profile, settings, etc.).
// - InRooms: map of RoomID to bool, indicating which rooms the user is in.
// Can be extended with app-scoped counters/badges if needed.
type AppUser struct {
	Nonce   uint64
	Data    []byte
	InRooms map[RoomID]bool
}

// -----------------------------------------------------------------------------
// Rooms & RoomUsers (app-defined shards)
// -----------------------------------------------------------------------------

// Room is an App-defined shard/topic (formerly "Channel").
// - Version:    CAS for structural changes (Join/Quit/Policy).
// - PolicyHash: hash of room policy (who must sign, per-asset caps/allowances, etc.).
// - Totals:     Σ RoomUser.Balances per asset (conservation check, fast audits).
// - UserRoot:   root of SMT keyed by UserID, leaf = HashRoomUser(row).
type Room struct {
	ID         RoomID
	App        AppID
	Version    uint64
	PolicyHash [32]byte
	Totals     map[AssetID]*Amount
	UserRoot   [32]byte
}

// RoomUser is the per-(Room,User) row.
// - RoomID:   identifies the room this user is in.
// - App:      identifies the app that created this room (for policy checks).
// - UserID:   identifies the user in this room.
// - Nonce:    increments on any update to this row (anti-replay).
// - DataHash: hash of room-scoped data (session/config/etc.).
// - Balances: per-asset amounts IN THIS ROOM (do not put Available here).
type RoomUser struct {
	RoomID   RoomID
	App      AppID
	UserID   UserID
	Nonce    uint64
	DataHash [32]byte
	Balances map[AssetID]*Amount
}
