---
sidebar_position: 3
title: Gaming with TicTacToe
description: Example of out to manage nitro state using go-nitro SDK
keywords: [erc7824, statechannels, nitro, sdk, development, state channels, ethereum scaling, L2]
tags:
  - erc7824
  - golang
  - go-nitro
  - nitro
  - docs
---
# TicTacToe Offchain

Below an example in go lang of usage of our upcoming go lang SDK
Typescript will be also available soon.

```go
package main

import (
 "errors"
 "testing"

 "github.com/ethereum/go-ethereum/common"
 ecrypto "github.com/ethereum/go-ethereum/crypto"
 "github.com/layer-3/clearsync/pkg/signer"
 "github.com/layer-3/neodax/internal/nitro"
 "github.com/stretchr/testify/require"
)

// TicTacToe represents a simple Nitro application implementing a Tic-Tac-Toe game.
type TicTacToe struct {
 players []signer.Signer
 grid    [3][3]byte // 3x3 grid for TicTacToe

 ch nitro.Channel
}

// NewTicTacToe initializes a new TicTacToe instance.
func NewTicTacToe(players []signer.Signer) *TicTacToe {
 fp := nitro.FixedPart{
  Participants: []common.Address{players[0].CommonAddress(), players[1].CommonAddress()},
 }
 vp := nitro.VariablePart{}
 s := nitro.StateFromFixedAndVariablePart(fp, vp)

 return &TicTacToe{
  players: players,
  grid:    [3][3]byte{},
  ch:      *nitro.NewChannel(s),
 }
}

// Definition returns the FixedPart of the Nitro state.
func (t *TicTacToe) Definition() nitro.FixedPart {
 return t.ch.FixedPart
}

// Data returns the VariablePart representing the current game state.
func (t *TicTacToe) Data(turn uint64) nitro.VariablePart {
 return nitro.VariablePart{
  AppData: encodeGrid(t.grid),
  Outcome: nitro.Exit{},
 }
}

func (t *TicTacToe) State(turn uint64) nitro.SignedState {
 return t.ch.SignedStateForTurnNum[turn]
}

// Validate checks if the current state is valid.
func (t *TicTacToe) Validate(turn uint64) bool {
 // Validate state can be done using the contract artifact
 return true
}

func (t *TicTacToe) LatestSupportedState() (nitro.SignedState, error) {
 return t.ch.LatestSupportedState()
}

// encodeGrid converts the grid to a byte slice for storage in AppData.
func encodeGrid(grid [3][3]byte) []byte {
 var data []byte
 // Encode using ABI
 for _, row := range grid {
  data = append(data, row[:]...)
 }
 return data
}

// MakeMove updates the game state with a player's move.
func (t *TicTacToe) MakeMove(player byte, x, y int) error {
 if x < 0 || x > 2 || y < 0 || y > 2 {
  return errors.New("invalid move: out of bounds")
 }
 if t.grid[x][y] != 0 {
  return errors.New("invalid move: cell already occupied")
 }
 t.grid[x][y] = player
 return nil
}

func (t *TicTacToe) StartGame() error {
 preFundState, err := t.ch.State(nitro.PreFundTurnNum)
 if err != nil {
  return err
 }

 for _, p := range t.players {
  if _, err := t.ch.SignAndAddState(preFundState.State(), p); err != nil {
   return err
  }
 }

 return nil
}

func (t *TicTacToe) FinishGame() error {
 lss, err := t.ch.LatestSignedState()
 if err != nil {
  return err
 }

 lastState := lss.State().Clone()
 lastState.IsFinal = true
 lastState.TurnNum += 1

 for _, p := range t.players {
  if _, err := t.ch.SignAndAddState(lastState, p); err != nil {
   return err
  }
 }

 return nil
}

func TestTicTacToe(t *testing.T) {
 // Create player signers
 var alice, bob signer.Signer

 app := NewTicTacToe([]signer.Signer{alice, bob})
 nch, err := nitro.NewClient() // Implements channel.Client
 require.NoError(t, err, "Error creating client")

 err = app.StartGame()
 require.NoError(t, err, "Error starting game")

 // Open the channel
 _, err = nch.Open(app)
 require.NoError(t, err, "Error opening channel")

 // Simulate moves
 err = app.MakeMove('X', 0, 0)
 require.NoError(t, err, "Invalid move")

 err = app.MakeMove('O', 1, 1)
 require.NoError(t, err, "Invalid move")

 err = app.FinishGame()
 require.NoError(t, err, "Error finishing game")

 // Close the channel at turn 2
 err = nch.Close(app)
 require.NoError(t, err, "Error closing channel")
}
```
