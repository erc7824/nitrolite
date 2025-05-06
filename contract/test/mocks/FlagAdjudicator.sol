// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Channel, State} from "../../src/interfaces/Types.sol";
import {IAdjudicator} from "../../src/interfaces/IAdjudicator.sol";
import {IComparable} from "../../src/interfaces/IComparable.sol";

contract FlagAdjudicator is IAdjudicator, IComparable {
    bool public adjudicateReturnValue = true;
    bool public compareReturnValue = true;

    function setAdjudicateReturnValue(bool value) external {
        adjudicateReturnValue = value;
    }

    function setCompareReturnValue(bool value) external {
        compareReturnValue = value;
    }

    function adjudicate(Channel calldata, State calldata, State[] calldata)
        external
        view
        override
        returns (bool valid)
    {
        // Always return true to indicate that the state is valid
        return adjudicateReturnValue;
    }

    function compare(State calldata, State calldata) external view override returns (int8) {
        return compareReturnValue ? int8(1) : -1;
    }
}
