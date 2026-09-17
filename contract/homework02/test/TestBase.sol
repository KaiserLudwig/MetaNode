// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @dev Foundry 作弊码接口（只声明用到的部分，避免依赖 forge-std）
interface Vm {
    function prank(address sender) external;
    function warp(uint256 newTimestamp) external;
    function deal(address account, uint256 newBalance) external;
}

address constant VM_ADDRESS = 0x7109709ECfa91a80626fF3989D68f67F5b1DD12D;

/// @dev 极简断言基类：不依赖 forge-std，clone 后可直接 forge test
abstract contract TestBase {
    Vm internal constant vm = Vm(VM_ADDRESS);

    function assertTrue(bool condition, string memory message) internal pure {
        require(condition, message);
    }

    function assertFalse(bool condition, string memory message) internal pure {
        require(!condition, message);
    }

    function assertEq(uint256 a, uint256 b, string memory message) internal pure {
        require(a == b, message);
    }

    function assertEq(bool a, bool b, string memory message) internal pure {
        require(a == b, message);
    }

    function assertEq(address a, address b, string memory message) internal pure {
        require(a == b, message);
    }
}
