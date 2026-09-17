// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {RomanNumeral} from "../src/RomanNumeral.sol";
import {TestBase} from "./TestBase.sol";

contract RomanNumeralTest is TestBase {
    RomanNumeral private roman;

    function setUp() public {
        roman = new RomanNumeral();
    }

    function test_IntToRomanLeetCodeExamples() public view {
        assertEq(roman.intToRoman(3), "III", "3 -> III");
        assertEq(roman.intToRoman(4), "IV", "4 -> IV");
        assertEq(roman.intToRoman(9), "IX", "9 -> IX");
        assertEq(roman.intToRoman(58), "LVIII", "58 -> LVIII");
        assertEq(roman.intToRoman(1994), "MCMXCIV", "1994 -> MCMXCIV");
    }

    function test_IntToRomanBoundaries() public view {
        assertEq(roman.intToRoman(1), "I", "1 -> I");
        assertEq(roman.intToRoman(3999), "MMMCMXCIX", "3999 -> MMMCMXCIX");
        assertEq(roman.intToRoman(3888), "MMMDCCCLXXXVIII", "3888 -> longest form");
    }

    function test_IntToRomanRevertsOutOfRange() public {
        try roman.intToRoman(0) {
            assertTrue(false, "0 should revert");
        } catch {
            assertTrue(true, "reverted as expected");
        }

        try roman.intToRoman(4000) {
            assertTrue(false, "4000 should revert");
        } catch {
            assertTrue(true, "reverted as expected");
        }
    }

    function test_RomanToIntLeetCodeExamples() public view {
        assertEq(roman.romanToInt("III"), 3, "III -> 3");
        assertEq(roman.romanToInt("IV"), 4, "IV -> 4");
        assertEq(roman.romanToInt("IX"), 9, "IX -> 9");
        assertEq(roman.romanToInt("LVIII"), 58, "LVIII -> 58");
        assertEq(roman.romanToInt("MCMXCIV"), 1994, "MCMXCIV -> 1994");
    }

    function test_RomanToIntBoundaries() public view {
        assertEq(roman.romanToInt("I"), 1, "I -> 1");
        assertEq(roman.romanToInt("MMMCMXCIX"), 3999, "MMMCMXCIX -> 3999");
    }

    function test_RomanToIntRejectsInvalidNumerals() public {
        try roman.romanToInt("IIII") {
            assertTrue(false, "IIII should revert");
        } catch {
            assertTrue(true, "reverted as expected");
        }

        try roman.romanToInt("ABC") {
            assertTrue(false, "illegal characters should revert");
        } catch {
            assertTrue(true, "reverted as expected");
        }

        try roman.romanToInt("") {
            assertTrue(false, "empty input should revert");
        } catch {
            assertTrue(true, "reverted as expected");
        }
    }

    function test_RoundTripAllValues() public view {
        for (uint256 n = 1; n <= 3999; n++) {
            assertEq(roman.romanToInt(roman.intToRoman(n)), n, "round trip should restore the number");
        }
    }
}
