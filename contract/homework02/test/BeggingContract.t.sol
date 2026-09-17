// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {BeggingContract} from "../src/BeggingContract.sol";
import {TestBase} from "./TestBase.sol";

contract BeggingContractTest is TestBase {
    BeggingContract private begging;

    address private alice = address(0xA11CE);
    address private bob = address(0xB0B);
    address private carol = address(0xCA401);
    address private dave = address(0xDA7E);

    /// @dev 测试合约需要能接收 withdraw 转出的 ETH
    receive() external payable {}

    function setUp() public {
        begging = new BeggingContract();
        vm.deal(alice, 100 ether);
        vm.deal(bob, 100 ether);
        vm.deal(carol, 100 ether);
        vm.deal(dave, 100 ether);
    }

    function test_OwnerIsDeployer() public view {
        assertEq(begging.owner(), address(this), "owner should be the deployer");
    }

    function test_InitialStateIsEmpty() public view {
        assertEq(begging.totalDonated(), 0, "totalDonated should start at 0");
        assertEq(begging.getDonation(alice), 0, "unknown donor should have 0");
        assertEq(begging.donorCount(), 0, "donor count should start at 0");
        assertTrue(begging.isDonationOpen(), "donation should be open by default");
    }

    function test_DonateRecordsAmount() public {
        vm.prank(alice);
        begging.donate{value: 1 ether}();

        assertEq(begging.getDonation(alice), 1 ether, "alice donation should be 1 ether");
        assertEq(begging.totalDonated(), 1 ether, "totalDonated should be 1 ether");
        assertEq(begging.donorCount(), 1, "should have 1 donor");
    }

    function test_MultipleDonationsAccumulate() public {
        vm.prank(alice);
        begging.donate{value: 1 ether}();
        vm.prank(alice);
        begging.donate{value: 2 ether}();

        assertEq(begging.getDonation(alice), 3 ether, "alice should have 3 ether in total");
        assertEq(begging.donorCount(), 1, "same donor must be counted once");
    }

    function test_DonationsAreTrackedPerAddress() public {
        vm.prank(alice);
        begging.donate{value: 1 ether}();
        vm.prank(bob);
        begging.donate{value: 5 ether}();

        assertEq(begging.getDonation(alice), 1 ether, "alice should have 1 ether");
        assertEq(begging.getDonation(bob), 5 ether, "bob should have 5 ether");
        assertEq(begging.totalDonated(), 6 ether, "total should be 6 ether");
        assertEq(address(begging).balance, 6 ether, "contract should hold 6 ether");
    }

    function test_PlainTransferCountsAsDonation() public {
        vm.prank(carol);
        (bool ok,) = address(begging).call{value: 2 ether}("");

        assertTrue(ok, "plain transfer should succeed");
        assertEq(begging.getDonation(carol), 2 ether, "plain transfer should be recorded");
    }

    function test_DonateZeroReverts() public {
        vm.prank(alice);
        (bool ok,) = address(begging).call{value: 0}(abi.encodeCall(BeggingContract.donate, ()));
        assertFalse(ok, "zero donation should revert");
    }

    function test_WithdrawTransfersEverythingToOwner() public {
        vm.prank(alice);
        begging.donate{value: 3 ether}();
        vm.prank(bob);
        begging.donate{value: 2 ether}();

        uint256 balanceBefore = address(this).balance;
        begging.withdraw();

        assertEq(address(begging).balance, 0, "contract should be drained");
        assertEq(address(this).balance, balanceBefore + 5 ether, "owner should receive 5 ether");
        assertEq(begging.getDonation(alice), 3 ether, "donation records should be kept");
        assertEq(begging.totalDonated(), 5 ether, "totalDonated should be kept");
    }

    function test_WithdrawCanBeCalledTwiceAfterNewDonation() public {
        vm.prank(alice);
        begging.donate{value: 1 ether}();
        begging.withdraw();

        vm.prank(bob);
        begging.donate{value: 2 ether}();
        begging.withdraw();

        assertEq(address(begging).balance, 0, "contract should be drained again");
        assertEq(begging.totalDonated(), 3 ether, "total should accumulate across withdrawals");
    }

    function test_WithdrawWithEmptyBalanceReverts() public {
        (bool ok,) = address(begging).call(abi.encodeCall(BeggingContract.withdraw, ()));
        assertFalse(ok, "withdraw without funds should revert");
    }

    function test_WithdrawRevertsForNonOwner() public {
        vm.prank(alice);
        begging.donate{value: 1 ether}();

        vm.prank(alice);
        (bool ok,) = address(begging).call(abi.encodeCall(BeggingContract.withdraw, ()));
        assertFalse(ok, "non-owner withdraw should revert");
        assertEq(address(begging).balance, 1 ether, "funds must stay in the contract");
    }

    function test_TopDonorsRanksThreeHighest() public {
        vm.prank(alice);
        begging.donate{value: 1 ether}();
        vm.prank(bob);
        begging.donate{value: 5 ether}();
        vm.prank(carol);
        begging.donate{value: 3 ether}();
        vm.prank(dave);
        begging.donate{value: 2 ether}();

        (address[3] memory who, uint256[3] memory amounts) = begging.topDonors();

        assertEq(who[0], bob, "1st should be bob");
        assertEq(amounts[0], 5 ether, "1st amount should be 5 ether");
        assertEq(who[1], carol, "2nd should be carol");
        assertEq(amounts[1], 3 ether, "2nd amount should be 3 ether");
        assertEq(who[2], dave, "3rd should be dave");
        assertEq(amounts[2], 2 ether, "3rd amount should be 2 ether");
    }

    function test_TopDonorsWithFewerThanThreeDonors() public {
        vm.prank(alice);
        begging.donate{value: 4 ether}();

        (address[3] memory who, uint256[3] memory amounts) = begging.topDonors();

        assertEq(who[0], alice, "1st should be alice");
        assertEq(amounts[0], 4 ether, "1st amount should be 4 ether");
        assertEq(amounts[1], 0, "2nd slot should be empty");
        assertEq(amounts[2], 0, "3rd slot should be empty");
    }

    function test_TopDonorsUsesAccumulatedAmounts() public {
        vm.prank(alice);
        begging.donate{value: 1 ether}();
        vm.prank(bob);
        begging.donate{value: 2 ether}();
        vm.prank(alice);
        begging.donate{value: 3 ether}();

        (address[3] memory who, uint256[3] memory amounts) = begging.topDonors();

        assertEq(who[0], alice, "alice total is 4 ether, should be 1st");
        assertEq(amounts[0], 4 ether, "1st amount should be 4 ether");
        assertEq(who[1], bob, "bob should be 2nd");
    }

    function test_DonationWindowBlocksEarlyDonation() public {
        uint256 now_ = block.timestamp;
        begging.setDonationWindow(now_ + 1 days, now_ + 2 days);

        assertFalse(begging.isDonationOpen(), "window should not be open yet");

        vm.prank(alice);
        (bool ok,) = address(begging).call{value: 1 ether}(abi.encodeCall(BeggingContract.donate, ()));
        assertFalse(ok, "donation before the window should revert");
    }

    function test_DonationWindowAllowsDonationInside() public {
        uint256 start = block.timestamp + 1 days;
        begging.setDonationWindow(start, start + 1 days);
        vm.warp(start + 1 hours);

        assertTrue(begging.isDonationOpen(), "window should be open");

        vm.prank(alice);
        begging.donate{value: 1 ether}();
        assertEq(begging.getDonation(alice), 1 ether, "donation inside the window should be recorded");
    }

    function test_DonationWindowBlocksLateDonation() public {
        uint256 start = block.timestamp + 1 days;
        begging.setDonationWindow(start, start + 1 days);
        vm.warp(start + 2 days);

        assertFalse(begging.isDonationOpen(), "window should be closed");

        vm.prank(alice);
        (bool ok,) = address(begging).call{value: 1 ether}(abi.encodeCall(BeggingContract.donate, ()));
        assertFalse(ok, "donation after the window should revert");
    }

    function test_SetDonationWindowIsOwnerOnly() public {
        vm.prank(alice);
        (bool ok,) = address(begging)
            .call(abi.encodeCall(BeggingContract.setDonationWindow, (block.timestamp, block.timestamp + 1 days)));
        assertFalse(ok, "non-owner should not be able to set the window");
    }

    function test_SetDonationWindowRejectsInvalidRange() public {
        (bool ok,) = address(begging)
            .call(
                abi.encodeCall(BeggingContract.setDonationWindow, (block.timestamp + 2 days, block.timestamp + 1 days))
            );
        assertFalse(ok, "end before start should revert");
    }
}
