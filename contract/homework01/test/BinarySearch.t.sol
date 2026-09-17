// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {BinarySearch} from "../src/BinarySearch.sol";
import {TestBase} from "./TestBase.sol";

contract BinarySearchTest is TestBase {
    BinarySearch private searcher;

    function setUp() public {
        searcher = new BinarySearch();
    }

    function test_FoundInMiddle() public view {
        uint256[] memory arr = _range(1, 7);
        (bool found, uint256 index) = searcher.search(arr, 4);

        assertTrue(found, "4 should be found");
        assertEq(index, 3, "4 is at index 3");
    }

    function test_FoundAtFirstPosition() public view {
        uint256[] memory arr = _range(1, 7);
        (bool found, uint256 index) = searcher.search(arr, 1);

        assertTrue(found, "1 should be found");
        assertEq(index, 0, "1 is at index 0");
    }

    function test_FoundAtLastPosition() public view {
        uint256[] memory arr = _range(1, 7);
        (bool found, uint256 index) = searcher.search(arr, 7);

        assertTrue(found, "7 should be found");
        assertEq(index, 6, "7 is at index 6");
    }

    function test_NotFoundBetweenElements() public view {
        uint256[] memory arr = _range(1, 7);
        (bool found,) = searcher.search(arr, 8);

        assertFalse(found, "8 should not be found");
    }

    function test_NotFoundInEmptyArray() public view {
        uint256[] memory arr = new uint256[](0);
        (bool found,) = searcher.search(arr, 1);

        assertFalse(found, "empty array should not contain anything");
    }

    function test_FoundInSingleElementArray() public view {
        uint256[] memory arr = new uint256[](1);
        arr[0] = 42;

        (bool found, uint256 index) = searcher.search(arr, 42);
        assertTrue(found, "42 should be found");
        assertEq(index, 0, "42 is at index 0");
    }

    function test_ReturnsMatchingIndexForDuplicates() public view {
        uint256[] memory arr = new uint256[](7);
        arr[0] = 1;
        arr[1] = 2;
        arr[2] = 2;
        arr[3] = 2;
        arr[4] = 3;
        arr[5] = 4;
        arr[6] = 5;

        (bool found, uint256 index) = searcher.search(arr, 2);
        assertTrue(found, "2 should be found");
        assertEq(arr[index], 2, "returned index must point at a matching element");
    }

    function test_SearchIndexReturnsMinusOneWhenMissing() public view {
        uint256[] memory arr = _range(1, 5);

        assertEq(searcher.searchIndex(arr, 3), int256(2), "3 is at index 2");
        assertEq(searcher.searchIndex(arr, 99), int256(-1), "missing target should return -1");
    }

    /// @dev 生成 [from, to] 的连续升序数组
    function _range(uint256 from, uint256 to) private pure returns (uint256[] memory) {
        uint256[] memory arr = new uint256[](to - from + 1);
        for (uint256 i = 0; i < arr.length; i++) {
            arr[i] = from + i;
        }
        return arr;
    }
}
