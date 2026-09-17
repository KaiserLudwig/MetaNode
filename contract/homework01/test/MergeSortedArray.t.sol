// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {MergeSortedArray} from "../src/MergeSortedArray.sol";
import {TestBase} from "./TestBase.sol";

contract MergeSortedArrayTest is TestBase {
    MergeSortedArray private merger;

    function setUp() public {
        merger = new MergeSortedArray();
    }

    function test_MergeLeetCodeExample() public view {
        uint256[] memory a = _toArray(new uint256[](3), 1, 2, 3);
        uint256[] memory b = _toArray(new uint256[](3), 2, 5, 6);
        uint256[] memory merged = merger.mergeSorted(a, b);

        _assertArrayEq(merged, _toArray(new uint256[](6), 1, 2, 2, 3, 5, 6), "should merge in ascending order");
    }

    function test_MergeWithEmptyLeft() public view {
        uint256[] memory a = new uint256[](0);
        uint256[] memory b = _toArray(new uint256[](2), 4, 7);
        uint256[] memory merged = merger.mergeSorted(a, b);

        _assertArrayEq(merged, _toArray(new uint256[](2), 4, 7), "empty left array should return right array");
    }

    function test_MergeWithEmptyRight() public view {
        uint256[] memory a = _toArray(new uint256[](2), 4, 7);
        uint256[] memory b = new uint256[](0);
        uint256[] memory merged = merger.mergeSorted(a, b);

        _assertArrayEq(merged, _toArray(new uint256[](2), 4, 7), "empty right array should return left array");
    }

    function test_MergeTwoEmptyArrays() public view {
        uint256[] memory merged = merger.mergeSorted(new uint256[](0), new uint256[](0));
        assertEq(merged.length, 0, "two empty arrays should produce empty array");
    }

    function test_MergeNonOverlappingRanges() public view {
        uint256[] memory a = _toArray(new uint256[](3), 1, 3, 5);
        uint256[] memory b = _toArray(new uint256[](3), 2, 4, 6);
        uint256[] memory merged = merger.mergeSorted(a, b);

        _assertArrayEq(merged, _toArray(new uint256[](6), 1, 2, 3, 4, 5, 6), "should interleave correctly");
    }

    function test_MergeKeepsDuplicates() public view {
        uint256[] memory a = _toArray(new uint256[](3), 1, 1, 2);
        uint256[] memory b = _toArray(new uint256[](3), 1, 2, 2);
        uint256[] memory merged = merger.mergeSorted(a, b);

        _assertArrayEq(merged, _toArray(new uint256[](6), 1, 1, 1, 2, 2, 2), "duplicates must be kept");
    }

    function _toArray(uint256[] memory arr, uint256 v0, uint256 v1, uint256 v2)
        private
        pure
        returns (uint256[] memory)
    {
        arr[0] = v0;
        arr[1] = v1;
        arr[2] = v2;
        return arr;
    }

    function _toArray(uint256[] memory arr, uint256 v0, uint256 v1, uint256 v2, uint256 v3, uint256 v4, uint256 v5)
        private
        pure
        returns (uint256[] memory)
    {
        arr[0] = v0;
        arr[1] = v1;
        arr[2] = v2;
        arr[3] = v3;
        arr[4] = v4;
        arr[5] = v5;
        return arr;
    }

    function _toArray(uint256[] memory arr, uint256 v0, uint256 v1) private pure returns (uint256[] memory) {
        arr[0] = v0;
        arr[1] = v1;
        return arr;
    }

    function _assertArrayEq(uint256[] memory actual, uint256[] memory expected, string memory message) private pure {
        require(actual.length == expected.length, message);
        for (uint256 i = 0; i < expected.length; i++) {
            require(actual[i] == expected[i], message);
        }
    }
}
