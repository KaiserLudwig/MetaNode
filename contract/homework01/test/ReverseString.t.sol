// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {ReverseString} from "../src/ReverseString.sol";
import {TestBase} from "./TestBase.sol";

contract ReverseStringTest is TestBase {
    ReverseString private utils;

    function setUp() public {
        utils = new ReverseString();
    }

    function test_ReverseExampleFromHomework() public view {
        assertEq(utils.reverse("abcde"), "edcba", "abcde should become edcba");
    }

    function test_ReverseEmptyString() public view {
        assertEq(utils.reverse(""), "", "empty string should stay empty");
    }

    function test_ReverseSingleCharacter() public view {
        assertEq(utils.reverse("a"), "a", "single character should stay the same");
    }

    function test_ReverseEvenLength() public view {
        assertEq(utils.reverse("abcd"), "dcba", "abcd should become dcba");
    }

    function test_ReversePalindromeIsUnchanged() public view {
        assertEq(utils.reverse("level"), "level", "palindrome should stay the same");
    }

    function test_ReverseTwiceRestoresOriginal() public view {
        string memory original = "hello world";
        assertEq(utils.reverse(utils.reverse(original)), original, "double reverse should restore the input");
    }
}
