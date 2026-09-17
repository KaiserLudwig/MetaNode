// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {NFTAuction} from "../NFTAuction.sol";

/// @title NFTAuctionV2 拍卖市场升级版本
/// @notice 演示 UUPS 升级：新增最小加价比例配置，并要求新出价至少高出该比例
contract NFTAuctionV2 is NFTAuction {
    /// @notice 最小加价比例（基点），0 表示不限制
    uint256 public minBidIncrementBps;

    event MinBidIncrementUpdated(uint256 bps);

    /// @notice 升级后初始化新状态变量（reinitializer(2) 保证只执行一次）
    function initializeV2(uint256 minBidIncrementBps_) external reinitializer(2) {
        require(minBidIncrementBps_ <= 5_000, "NFTAuctionV2: increment too large");
        minBidIncrementBps = minBidIncrementBps_;
        emit MinBidIncrementUpdated(minBidIncrementBps_);
    }

    /// @notice 平台 owner 调整最小加价比例
    function setMinBidIncrement(uint256 bps) external onlyOwner {
        require(bps <= 5_000, "NFTAuctionV2: increment too large");
        minBidIncrementBps = bps;
        emit MinBidIncrementUpdated(bps);
    }

    function version() external pure override returns (string memory) {
        return "2.0.0";
    }

    /// @dev V2 的新规则：再次出价必须比当前最高价高出 minBidIncrementBps
    function _validateBid(Auction storage auction, uint256 amount) internal view override {
        if (auction.highestBid > 0 && minBidIncrementBps > 0) {
            uint256 minimum = auction.highestBid + (auction.highestBid * minBidIncrementBps) / BPS_DENOMINATOR;
            require(amount >= minimum, "NFTAuctionV2: bid increment too small");
        }
    }
}
