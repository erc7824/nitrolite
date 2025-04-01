// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {IAdjudicator} from "../interfaces/IAdjudicator.sol";
import {Channel, State, Allocation, Signature} from "../interfaces/Types.sol";
import {Utils} from "../Utils.sol";

/**
 * @title TicTacToe Adjudicator
 * @notice An adjudicator that implements a Tic-Tac-Toe game on a 3x3 grid.
 * @dev Host plays as 'X' and Guest plays as 'O'. Host plays first.
 * Each move must be signed by the current player.
 * When a player gets three in a row, column, or diagonal, the game ends with FINAL status.
 */
contract TicTacToe is IAdjudicator {
    /// @notice Error thrown when signature verification fails
    error InvalidSignature();
    /// @notice Error thrown when turn order is violated
    error InvalidTurn();
    /// @notice Error thrown when a position is already taken
    error InvalidPosition();
    /// @notice Error thrown when a position is out of bounds
    error OutOfBounds();
    /// @notice Error thrown when insufficient signatures are provided
    error InsufficientSignatures();

    uint256 private constant HOST = 0;
    uint256 private constant GUEST = 1;
    uint256 private constant EMPTY = 0;
    uint256 private constant X = 1; // Host
    uint256 private constant O = 2; // Guest

    // GameGrid encodes the current state of the Tic-Tac-Toe board
    struct GameGrid {
        // Grid is represented as a 3x3 array
        // 0 = empty, 1 = X (Host), 2 = O (Guest)
        uint256[3][3] grid;
        // Total number of moves played
        uint256 moveCount;
    }

    /**
     * @notice Validates that moves in Tic-Tac-Toe follow the rules of the game
     * @param chan The channel configuration
     * @param candidate The proposed game state
     * @param proofs Array containing the previous state signed by the previous participant
     * @return decision The status of the channel after adjudication
     */
    function adjudicate(Channel calldata chan, State calldata candidate, State[] calldata proofs)
        external
        pure
        override
        returns (Status decision)
    {

        // Check if we have at least one signature
        if (candidate.sigs.length == 0) return Status.INVALID;

        // Get the state hash for signature verification
        bytes32 stateHash = Utils.getStateHash(chan, candidate);

        // Decode the game grid from candidate state.data
        GameGrid memory candidateGrid = abi.decode(candidate.data, (GameGrid));

        // INITIAL STATE ACTIVATION: No proofs provided
        if (proofs.length == 0) {
            // First signature must be from HOST who initializes the game
            if (!Utils.verifySignature(stateHash, candidate.sigs[0], chan.participants[HOST])) {
                return Status.VOID;
            }

            // If only Host has signed, channel is PARTIAL
            if (candidate.sigs.length < 2) {
                return Status.PARTIAL;
            }

            // Both signatures provided, verify Guest's signature
            if (!Utils.verifySignature(stateHash, candidate.sigs[1], chan.participants[GUEST])) {
                return Status.VOID;
            }

            // Verify initial grid is empty and move count is 0
            if (!isGridEmpty(candidateGrid)) {
                return Status.INVALID;
            }

            // Channel becomes ACTIVE with an empty grid
            return Status.ACTIVE;
        }

        // NORMAL STATE TRANSITION: Proof provided.
        // Ensure proof state has at least one signature
        if (proofs[0].sigs.length == 0) return Status.INVALID;

        GameGrid memory previousGrid = abi.decode(proofs[0].data, (GameGrid));
        bytes32 proofStateHash = Utils.getStateHash(chan, proofs[0]);

        // Verify the move is valid (check number of moves, grid changes, etc.)
        if (!isValidMove(previousGrid, candidateGrid)) {
            return Status.INVALID;
        }

        // Check whose turn it is based on the move count in previous grid
        bool isHostTurn = previousGrid.moveCount % 2 == 0;
        address currentPlayer = isHostTurn ? chan.participants[HOST] : chan.participants[GUEST];
        address previousPlayer = isHostTurn ? chan.participants[GUEST] : chan.participants[HOST];

        // Verify the current player has signed the candidate state
        if (!Utils.verifySignature(stateHash, candidate.sigs[0], currentPlayer)) {
            return Status.INVALID;
        }

        // Verify the previous player has signed the proof state
        if (!Utils.verifySignature(proofStateHash, proofs[0].sigs[0], previousPlayer)) {
            return Status.INVALID;
        }

        // Check if the game is over (win or draw)
        uint256 winner = checkWinner(candidateGrid);
        if (winner == X || winner == O || isDraw(candidateGrid)) {
            return Status.FINAL;
        }

        // Valid state transition, channel remains ACTIVE
        return Status.ACTIVE;
    }

    /**
     * @notice Checks if the grid is empty (initial state)
     * @param grid The game grid to check
     * @return True if the grid is empty and move count is 0
     */
    function isGridEmpty(GameGrid memory grid) internal pure returns (bool) {
        if (grid.moveCount != 0) return false;

        for (uint256 i = 0; i < 3; i++) {
            for (uint256 j = 0; j < 3; j++) {
                if (grid.grid[i][j] != EMPTY) return false;
            }
        }

        return true;
    }

    /**
     * @notice Checks if the move from previousGrid to candidateGrid is valid
     * @param previousGrid The previous game state
     * @param candidateGrid The new game state
     * @return True if the move is valid
     */
    function isValidMove(GameGrid memory previousGrid, GameGrid memory candidateGrid) internal pure returns (bool) {
        // Move count should increment by exactly 1
        if (candidateGrid.moveCount != previousGrid.moveCount + 1) {
            return false;
        }

        // Determine which player's turn it is based on move count
        uint256 currentPlayer = previousGrid.moveCount % 2 == 0 ? X : O;

        // Count changes between grids - should be exactly one change
        uint256 changesCount = 0;
        uint256 row;
        uint256 col;

        for (uint256 i = 0; i < 3; i++) {
            for (uint256 j = 0; j < 3; j++) {
                if (previousGrid.grid[i][j] != candidateGrid.grid[i][j]) {
                    changesCount++;
                    row = i;
                    col = j;

                    // The cell should have been empty and now contain the current player's marker
                    if (previousGrid.grid[i][j] != EMPTY || candidateGrid.grid[i][j] != currentPlayer) {
                        return false;
                    }
                }
            }
        }

        // Ensure exactly one change occurred
        return changesCount == 1;
    }

    /**
     * @notice Checks if there's a winner in the current grid
     * @param grid The game grid to check
     * @return The winner (X=1, O=2) or EMPTY=0 if no winner
     */
    function checkWinner(GameGrid memory grid) internal pure returns (uint256) {
        // Check rows
        for (uint256 i = 0; i < 3; i++) {
            if (grid.grid[i][0] != EMPTY && grid.grid[i][0] == grid.grid[i][1] && grid.grid[i][1] == grid.grid[i][2]) {
                return grid.grid[i][0];
            }
        }

        // Check columns
        for (uint256 j = 0; j < 3; j++) {
            if (grid.grid[0][j] != EMPTY && grid.grid[0][j] == grid.grid[1][j] && grid.grid[1][j] == grid.grid[2][j]) {
                return grid.grid[0][j];
            }
        }

        // Check diagonal top-left to bottom-right
        if (grid.grid[0][0] != EMPTY && grid.grid[0][0] == grid.grid[1][1] && grid.grid[1][1] == grid.grid[2][2]) {
            return grid.grid[0][0];
        }

        // Check diagonal top-right to bottom-left
        if (grid.grid[0][2] != EMPTY && grid.grid[0][2] == grid.grid[1][1] && grid.grid[1][1] == grid.grid[2][0]) {
            return grid.grid[0][2];
        }

        return EMPTY; // No winner
    }

    /**
     * @notice Checks if the game is a draw (all cells filled, no winner)
     * @param grid The game grid to check
     * @return True if the game is a draw
     */
    function isDraw(GameGrid memory grid) internal pure returns (bool) {
        // If there's a winner, it's not a draw
        if (checkWinner(grid) != EMPTY) {
            return false;
        }

        // If all cells are filled, it's a draw
        return grid.moveCount == 9;
    }
}
