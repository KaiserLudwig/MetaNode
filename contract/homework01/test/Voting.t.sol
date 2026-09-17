// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Voting} from "../src/Voting.sol";
import {TestBase} from "./TestBase.sol";

contract VotingTest is TestBase {
    Voting private voting;

    function setUp() public {
        voting = new Voting();
    }

    function test_UnknownCandidateHasZeroVotes() public view {
        assertEq(voting.getVotes("alice"), 0, "unknown candidate should have 0 votes");
    }

    function test_VoteIncrementsVotes() public {
        voting.vote("alice");
        assertEq(voting.getVotes("alice"), 1, "alice should have 1 vote");

        voting.vote("alice");
        assertEq(voting.getVotes("alice"), 2, "alice should have 2 votes");
    }

    function test_VotesAreTrackedPerCandidate() public {
        voting.vote("alice");
        voting.vote("bob");
        voting.vote("bob");

        assertEq(voting.getVotes("alice"), 1, "alice should have 1 vote");
        assertEq(voting.getVotes("bob"), 2, "bob should have 2 votes");
    }

    function test_CandidatesAreRegisteredOnce() public {
        voting.vote("alice");
        voting.vote("alice");
        voting.vote("bob");

        assertEq(voting.candidateCount(), 2, "should register 2 candidates");
        string[] memory list = voting.candidates();
        assertEq(list[0], "alice", "first candidate should be alice");
        assertEq(list[1], "bob", "second candidate should be bob");
    }

    function test_ResetVotesClearsAllCandidates() public {
        voting.vote("alice");
        voting.vote("alice");
        voting.vote("bob");

        voting.resetVotes();

        assertEq(voting.getVotes("alice"), 0, "alice votes should be reset");
        assertEq(voting.getVotes("bob"), 0, "bob votes should be reset");
        assertEq(voting.candidateCount(), 2, "candidates should stay registered");
    }

    function test_VoteAfterResetStartsFromZero() public {
        voting.vote("alice");
        voting.resetVotes();
        voting.vote("alice");

        assertEq(voting.getVotes("alice"), 1, "vote after reset should count from 0");
    }

    function test_VoteRevertsOnEmptyCandidate() public {
        try voting.vote("") {
            assertTrue(false, "empty candidate should revert");
        } catch {
            assertTrue(true, "reverted as expected");
        }
    }
}
