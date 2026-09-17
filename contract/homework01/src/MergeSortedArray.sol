// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title MergeSortedArray 合并两个有序数组
/// @notice 作业 5：将两个升序数组合并为一个升序数组
contract MergeSortedArray {
    /// @notice 合并两个升序数组，返回新的升序数组
    /// @dev 双指针法，时间复杂度 O(m + n)
    function mergeSorted(uint256[] memory a, uint256[] memory b) public pure returns (uint256[] memory) {
        uint256[] memory merged = new uint256[](a.length + b.length);
        uint256 i = 0;
        uint256 j = 0;
        uint256 k = 0;

        while (i < a.length && j < b.length) {
            if (a[i] <= b[j]) {
                merged[k] = a[i];
                i++;
            } else {
                merged[k] = b[j];
                j++;
            }
            k++;
        }

        while (i < a.length) {
            merged[k] = a[i];
            i++;
            k++;
        }
        while (j < b.length) {
            merged[k] = b[j];
            j++;
            k++;
        }

        return merged;
    }
}
