// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Channel, State} from "../../src/interfaces/Types.sol";
import {IAdjudicator} from "../../src/interfaces/IAdjudicator.sol";
import {IComparable} from "../../src/interfaces/IComparable.sol";

contract FlagAdjudicator is IAdjudicator, IComparable {
    bool public flag;

    constructor(bool flag_) {
        flag = flag_;
    }

    function setFlag(bool _flag) external {
        flag = _flag;
    }

    function adjudicate(Channel calldata, State calldata, State[] calldata)
        external
        view
        override
        returns (bool valid)
    {
        // Always return true to indicate that the state is valid
        return flag;
    }

    function compare(State calldata, State calldata) external pure override returns (int8) {
        // Always return 1 to indicate the candidate state is newer
        return 1;
    }
}
