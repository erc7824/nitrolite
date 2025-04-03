// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Channel, State} from "../../src/interfaces/Types.sol";
import {IAdjudicator} from "../../src/interfaces/IAdjudicator.sol";

contract FlagAdjudicator is IAdjudicator {
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

}
