// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title RomanNumeral 整数与罗马数字互转
/// @notice 作业 3 / 4：整数转罗马数字，罗马数字转整数（支持 1 ~ 3999）
contract RomanNumeral {
    /// @notice 整数转罗马数字，例如 1994 -> "MCMXCIV"
    /// @dev 贪心法：从大到小反复扣除对应数值
    function intToRoman(uint256 num) public pure returns (string memory) {
        require(num > 0 && num <= 3999, "RomanNumeral: out of range 1..3999");

        uint256[13] memory values = [uint256(1000), 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1];
        string[13] memory symbols = ["M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"];

        // 1..3999 内最长的罗马数字是 3888 = MMMDCCCLXXXVIII，共 15 个字符
        bytes memory buffer = new bytes(20);
        uint256 length = 0;

        for (uint256 i = 0; i < 13; i++) {
            while (num >= values[i]) {
                num -= values[i];
                bytes memory symbol = bytes(symbols[i]);
                for (uint256 j = 0; j < symbol.length; j++) {
                    buffer[length] = symbol[j];
                    length++;
                }
            }
        }

        bytes memory result = new bytes(length);
        for (uint256 k = 0; k < length; k++) {
            result[k] = buffer[k];
        }
        return string(result);
    }

    /// @notice 罗马数字转整数，例如 "MCMXCIV" -> 1994
    /// @dev 从右向左扫描：比右侧符号小则减，否则加；最后用 intToRoman 回环校验，
    ///      因此 "IIII"、"IIV" 这类非标准写法会被拒绝
    function romanToInt(string memory roman) public pure returns (uint256) {
        bytes memory chars = bytes(roman);
        uint256 length = chars.length;
        require(length > 0, "RomanNumeral: empty input");
        require(length <= 15, "RomanNumeral: too long");

        uint256 total = 0;
        uint256 prev = 0;

        for (uint256 i = length; i > 0; i--) {
            uint256 cur = _valueOf(chars[i - 1]);
            if (cur < prev) {
                total -= cur;
            } else {
                total += cur;
                prev = cur;
            }
        }

        require(keccak256(bytes(intToRoman(total))) == keccak256(chars), "RomanNumeral: invalid numeral");
        return total;
    }

    /// @dev 单个罗马字符对应的数值，非法字符直接 revert
    function _valueOf(bytes1 char) private pure returns (uint256) {
        if (char == "I") return 1;
        if (char == "V") return 5;
        if (char == "X") return 10;
        if (char == "L") return 50;
        if (char == "C") return 100;
        if (char == "D") return 500;
        if (char == "M") return 1000;
        // 该函数只在 romanToInt 的循环里被调用，非法字符必须立即中断解析
        // forge-lint: disable-next-line(require-revert-in-loop)
        revert("RomanNumeral: illegal character");
    }
}
