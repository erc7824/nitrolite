// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Channel, State} from "../../src/interfaces/Types.sol";
import {IAdjudicator} from "../../src/interfaces/IAdjudicator.sol";
import {IComparable} from "../../src/interfaces/IComparable.sol";

contract FlagAdjudicator is IAdjudicator, IComparable {
    bool public adjudicateReturnValue = true;
    int8 public compareReturnValue = 1;

    function setAdjudicateReturnValue(bool value) external {
        adjudicateReturnValue = value;
    }

    function setCompareReturnValue(int8 value) external {
        compareReturnValue = value;
    }

    function adjudicate(Channel calldata, State calldata, State[] calldata) external view returns (bool valid) {
        return adjudicateReturnValue;
    }

    function compare(State calldata, State calldata) external view virtual returns (int8) {
        return compareReturnValue;
    }
}
