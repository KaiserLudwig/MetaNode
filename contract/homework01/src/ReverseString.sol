// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title ReverseString 字符串反转
/// @notice 作业 2：输入 "abcde" 输出 "edcba"
contract ReverseString {
    /// @notice 反转字符串
    /// @dev Solidity 的 string 底层是 UTF-8 字节序列，这里是按字节反转，
    ///      对 ASCII 完全正确；对多字节字符（中文、emoji）会拆散字节，属于有意保留的简化
    function reverse(string memory s) public pure returns (string memory) {
        bytes memory original = bytes(s);
        uint256 length = original.length;
        bytes memory reversed = new bytes(length);

        for (uint256 i = 0; i < length; i++) {
            reversed[i] = original[length - 1 - i];
        }

        return string(reversed);
    }
}
