// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

/**
 * @title WadMath
 * @notice Library for converting amounts to WAD (18 decimals) format using method syntax
 * @dev Provides toWad() as attached functions for uint256 and int256 types
 */
library WadMath {
    uint8 constant MAX_PRECISION = 18;

    /**
     * @notice Converts a uint256 amount to WAD (18 decimals)
     * @param amount The amount to convert
     * @param decimals The current decimals of the amount
     * @return The amount scaled to 18 decimals
     */
    function toWad(uint256 amount, uint8 decimals) internal pure returns (uint256) {
        if (decimals > MAX_PRECISION) {
            revert("decimals exceed max precision");
        }

        if (decimals == MAX_PRECISION) {
            return amount;
        }

        return amount * 10 ** (MAX_PRECISION - decimals);
    }

    /**
     * @notice Converts an int256 amount to WAD (18 decimals)
     * @param amount The amount to convert
     * @param decimals The current decimals of the amount
     * @return The amount scaled to 18 decimals
     */
    function toWad(int256 amount, uint8 decimals) internal pure returns (int256) {
        if (decimals > MAX_PRECISION) {
            revert("decimals exceed max precision");
        }

        if (decimals == MAX_PRECISION) {
            return amount;
        }

        return amount * int256(10 ** (MAX_PRECISION - decimals));
    }
}
