// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title BinarySearch 二分查找
/// @notice 作业 6：在升序数组中查找目标值
contract BinarySearch {
    /// @notice 在升序数组中查找 target
    /// @return found 是否找到
    /// @return index 找到时为目标下标，未找到时为 0（配合 found 使用）
    /// @dev 时间复杂度 O(log n)；数组含重复元素时返回其中任意一个匹配下标
    function search(uint256[] memory arr, uint256 target) public pure returns (bool found, uint256 index) {
        uint256 low = 0;
        uint256 high = arr.length;

        while (low < high) {
            uint256 mid = low + (high - low) / 2;
            if (arr[mid] == target) {
                found = true;
                index = mid;
                return (found, index);
            }
            if (arr[mid] < target) {
                low = mid + 1;
            } else {
                high = mid;
            }
        }

        return (found, index);
    }

    /// @notice 与 search 等价，但未找到时返回 -1，便于直接判断
    function searchIndex(uint256[] memory arr, uint256 target) public pure returns (int256) {
        (bool found, uint256 index) = search(arr, target);
        // 数组长度不可能超过 int256 上限，转换是安全的
        // forge-lint: disable-next-line(unsafe-typecast)
        return found ? int256(index) : int256(-1);
    }
}
