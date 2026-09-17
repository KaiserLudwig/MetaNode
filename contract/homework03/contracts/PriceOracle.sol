// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {AggregatorV3Interface} from "@chainlink/contracts/src/v0.8/shared/interfaces/AggregatorV3Interface.sol";
import {IERC20Metadata} from "@openzeppelin/contracts/token/ERC20/extensions/IERC20Metadata.sol";
import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";

/// @title PriceOracle Chainlink 价格预言机封装
/// @notice 读取 Chainlink price feed，把 ETH / ERC20 的数量换算成美元
contract PriceOracle is Ownable {
    /// @notice 所有价格统一用 8 位小数表示（与 Chainlink USD feed 一致）
    uint256 public constant PRICE_DECIMALS = 8;

    /// @notice 价格过期时间，超过则视为无效
    uint256 public constant MAX_STALENESS = 1 hours;

    /// @notice 代币地址 => Chainlink feed；address(0) 代表原生 ETH
    mapping(address => AggregatorV3Interface) public priceFeeds;

    /// @param token 代币地址（address(0) 表示 ETH）
    /// @param feed 对应的 Chainlink feed
    event PriceFeedUpdated(address indexed token, address indexed feed);

    constructor(address initialOwner) Ownable(initialOwner) {}

    /// @notice 设置/更新某个代币的 Chainlink feed
    function setPriceFeed(address token, address feed) external onlyOwner {
        require(feed != address(0), "PriceOracle: feed is zero");
        priceFeeds[token] = AggregatorV3Interface(feed);
        emit PriceFeedUpdated(token, feed);
    }

    /// @notice 读取某个代币的美元价格（8 位小数）
    function getPrice(address token) public view returns (uint256 price) {
        AggregatorV3Interface feed = priceFeeds[token];
        require(address(feed) != address(0), "PriceOracle: feed not set");

        (uint80 roundId, int256 answer,, uint256 updatedAt, uint80 answeredInRound) = feed.latestRoundData();
        require(answer > 0, "PriceOracle: invalid price");
        require(updatedAt != 0 && block.timestamp - updatedAt <= MAX_STALENESS, "PriceOracle: stale price");
        require(answeredInRound >= roundId, "PriceOracle: stale round");

        return scaleTo8(uint256(answer), feed.decimals());
    }

    /// @notice 把某个代币的数量换算成美元（8 位小数）
    /// @param token 代币地址（address(0) 表示 ETH）
    /// @param amount 以该代币最小单位表示的数量
    function convertToUsd(address token, uint256 amount) public view returns (uint256 usdValue) {
        uint256 price = getPrice(token);
        uint8 tokenDecimals = decimalsOf(token);
        return (amount * price) / (10 ** tokenDecimals);
    }

    /// @notice 代币精度：ETH 固定 18，ERC20 读取 decimals()
    function decimalsOf(address token) public view returns (uint8) {
        if (token == address(0)) {
            return 18;
        }
        return IERC20Metadata(token).decimals();
    }

    /// @dev 把任意精度的价格统一成 8 位小数
    function scaleTo8(uint256 value, uint8 feedDecimals) internal pure returns (uint256) {
        if (feedDecimals == PRICE_DECIMALS) {
            return value;
        }
        if (feedDecimals < PRICE_DECIMALS) {
            return value * (10 ** (PRICE_DECIMALS - feedDecimals));
        }
        return value / (10 ** (feedDecimals - PRICE_DECIMALS));
    }
}
