// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @dev 极简断言基类：不依赖 forge-std，clone 后可直接 forge test
abstract contract TestBase {
    function assertTrue(bool condition, string memory message) internal pure {
        require(condition, message);
    }

    function assertFalse(bool condition, string memory message) internal pure {
        require(!condition, message);
    }

    function assertEq(uint256 a, uint256 b, string memory message) internal pure {
        require(a == b, message);
    }

    function assertEq(int256 a, int256 b, string memory message) internal pure {
        require(a == b, message);
    }

    function assertEq(bool a, bool b, string memory message) internal pure {
        require(a == b, message);
    }

    function assertEq(address a, address b, string memory message) internal pure {
        require(a == b, message);
    }

    function assertEq(string memory a, string memory b, string memory message) internal pure {
        require(keccak256(bytes(a)) == keccak256(bytes(b)), message);
    }
}
