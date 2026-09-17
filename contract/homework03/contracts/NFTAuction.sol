// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Initializable} from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import {UUPSUpgradeable} from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import {OwnableUpgradeable} from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import {IERC721} from "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
// OZ v5.5 起不再提供 ReentrancyGuardUpgradeable：普通版带构造函数会被升级插件拒绝，
// 而瞬态存储版（EIP-1153 / Cancun）无构造函数且 @custom:stateless，是代理合约的推荐选择
import {ReentrancyGuardTransient} from "@openzeppelin/contracts/utils/ReentrancyGuardTransient.sol";
import {PriceOracle} from "./PriceOracle.sol";

/// @title NFTAuction NFT 拍卖市场
/// @notice 支持 ETH / ERC20 出价，成交价通过 Chainlink 预言机换算成美元并按档位收取动态手续费；
///         合约采用 UUPS 代理模式，可通过 scripts/upgrade.js 升级到 NFTAuctionV2
contract NFTAuction is Initializable, OwnableUpgradeable, ReentrancyGuardTransient, UUPSUpgradeable {
    using SafeERC20 for IERC20;

    enum AuctionState {
        Created,
        Ended,
        Cancelled
    }

    struct Auction {
        address seller;
        address nft;
        uint256 tokenId;
        address paymentToken; // address(0) 表示 ETH
        uint256 startingPrice;
        uint256 highestBid;
        address highestBidder;
        uint64 endTime;
        AuctionState state;
    }

    uint256 public constant BPS_DENOMINATOR = 10_000;
    uint256 public constant MIN_DURATION = 1 minutes;
    uint256 public constant MAX_DURATION = 30 days;

    // ---- 动态手续费档位（额外挑战）----
    uint256 public constant FEE_LOW_BPS = 50; // 成交价 < 100 USD    → 0.5%
    uint256 public constant FEE_MID_BPS = 100; // 100 ~ 1000 USD     → 1%
    uint256 public constant FEE_HIGH_BPS = 200; // >= 1000 USD       → 2%
    uint256 public constant FEE_TIER1_LIMIT_USD = 100 * 1e8; // 100 USD（8 位小数）
    uint256 public constant FEE_TIER2_LIMIT_USD = 1000 * 1e8; // 1000 USD

    /// @notice Chainlink 价格预言机
    PriceOracle public priceOracle;

    /// @notice 手续费收款地址
    address public feeRecipient;

    /// @notice 拍卖自增编号
    uint256 public auctionCount;

    mapping(uint256 => Auction) private _auctions;

    /// @param auctionId 拍卖编号
    /// @param seller 卖家
    /// @param nft NFT 合约
    /// @param tokenId NFT 编号
    /// @param paymentToken 支付代币（address(0) 表示 ETH）
    /// @param startingPrice 起拍价
    /// @param endTime 结束时间
    event AuctionCreated(
        uint256 indexed auctionId,
        address indexed seller,
        address indexed nft,
        uint256 tokenId,
        address paymentToken,
        uint256 startingPrice,
        uint64 endTime
    );

    /// @param bidder 出价人
    /// @param amount 出价金额
    /// @param usdValue 出价对应的美元估值（8 位小数，取价失败时为 0）
    event BidPlaced(uint256 indexed auctionId, address indexed bidder, uint256 amount, uint256 usdValue);

    event AuctionEnded(uint256 indexed auctionId, address indexed winner, uint256 amount, uint256 fee);
    event AuctionCancelled(uint256 indexed auctionId);
    event PriceOracleUpdated(address indexed oracle);
    event FeeRecipientUpdated(address indexed feeRecipient);

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    /// @notice 初始化（代理模式下代替构造函数）
    function initialize(address initialOwner, address oracle, address feeRecipient_) public initializer {
        __Ownable_init(initialOwner);

        priceOracle = PriceOracle(oracle);
        feeRecipient = feeRecipient_;
    }

    /// @notice 合约版本，升级到 V2 后会变成 2.0.0
    function version() external pure virtual returns (string memory) {
        return "1.0.0";
    }

    // ---------------------------------------------------------------- 拍卖

    /// @notice 创建拍卖：NFT 会先转入本合约托管
    function createAuction(
        address nft,
        uint256 tokenId,
        address paymentToken,
        uint256 startingPrice,
        uint256 duration
    ) external nonReentrant returns (uint256 auctionId) {
        require(duration >= MIN_DURATION && duration <= MAX_DURATION, "NFTAuction: invalid duration");
        require(startingPrice > 0, "NFTAuction: invalid starting price");
        require(IERC721(nft).ownerOf(tokenId) == msg.sender, "NFTAuction: not the token owner");
        require(
            IERC721(nft).isApprovedForAll(msg.sender, address(this))
                || IERC721(nft).getApproved(tokenId) == address(this),
            "NFTAuction: not approved"
        );

        IERC721(nft).transferFrom(msg.sender, address(this), tokenId);

        auctionId = ++auctionCount;
        uint64 endTime = uint64(block.timestamp + duration);

        _auctions[auctionId] = Auction({
            seller: msg.sender,
            nft: nft,
            tokenId: tokenId,
            paymentToken: paymentToken,
            startingPrice: startingPrice,
            highestBid: 0,
            highestBidder: address(0),
            endTime: endTime,
            state: AuctionState.Created
        });

        emit AuctionCreated(auctionId, msg.sender, nft, tokenId, paymentToken, startingPrice, endTime);
    }

    /// @notice 用 ETH 出价
    function bidEth(uint256 auctionId) external payable nonReentrant {
        Auction storage auction = _activeAuction(auctionId);
        require(auction.paymentToken == address(0), "NFTAuction: not an ETH auction");
        _placeBid(auctionId, auction, msg.value);
    }

    /// @notice 用 ERC20 出价，代币会先转入本合约托管
    function bidErc20(uint256 auctionId, uint256 amount) external nonReentrant {
        Auction storage auction = _activeAuction(auctionId);
        require(auction.paymentToken != address(0), "NFTAuction: not an ERC20 auction");

        IERC20(auction.paymentToken).safeTransferFrom(msg.sender, address(this), amount);
        _placeBid(auctionId, auction, amount);
    }

    /// @notice 结束拍卖：NFT 归最高出价者，资金扣除手续费后归卖家
    function endAuction(uint256 auctionId) external nonReentrant {
        Auction storage auction = _auctions[auctionId];
        require(auction.seller != address(0), "NFTAuction: unknown auction");
        require(auction.state == AuctionState.Created, "NFTAuction: auction not active");
        require(block.timestamp >= auction.endTime, "NFTAuction: auction still running");

        auction.state = AuctionState.Ended;

        address winner = auction.highestBidder;

        // 无人出价：NFT 退回卖家
        if (winner == address(0)) {
            IERC721(auction.nft).transferFrom(address(this), auction.seller, auction.tokenId);
            emit AuctionEnded(auctionId, address(0), 0, 0);
            return;
        }

        uint256 amount = auction.highestBid;
        (uint256 fee,) = _calculateFee(auction.paymentToken, amount);
        if (feeRecipient == address(0)) {
            fee = 0;
        }

        IERC721(auction.nft).transferFrom(address(this), winner, auction.tokenId);

        if (fee > 0) {
            _payout(auction.paymentToken, feeRecipient, fee);
        }
        _payout(auction.paymentToken, auction.seller, amount - fee);

        emit AuctionEnded(auctionId, winner, amount, fee);
    }

    /// @notice 取消拍卖：仅卖家或平台 owner 可调用，且必须无人出价
    function cancelAuction(uint256 auctionId) external nonReentrant {
        Auction storage auction = _auctions[auctionId];
        require(auction.seller != address(0), "NFTAuction: unknown auction");
        require(auction.state == AuctionState.Created, "NFTAuction: auction not active");
        require(msg.sender == auction.seller || msg.sender == owner(), "NFTAuction: not allowed");
        require(auction.highestBidder == address(0), "NFTAuction: already has bids");

        auction.state = AuctionState.Cancelled;
        IERC721(auction.nft).transferFrom(address(this), auction.seller, auction.tokenId);

        emit AuctionCancelled(auctionId);
    }

    // ---------------------------------------------------------------- 查询

    /// @notice 读取拍卖详情
    function getAuction(uint256 auctionId) external view returns (Auction memory) {
        return _auctions[auctionId];
    }

    /// @notice 把任意代币数量换算成美元（8 位小数），取价失败时 available = false
    function quoteUsd(address token, uint256 amount) public view returns (uint256 usdValue, bool available) {
        if (address(priceOracle) == address(0)) {
            return (0, false);
        }
        try priceOracle.convertToUsd(token, amount) returns (uint256 value) {
            return (value, true);
        } catch {
            return (0, false);
        }
    }

    /// @notice 当前最高出价的美元估值，方便前端对比 ETH 与 ERC20 出价
    function highestBidUsd(uint256 auctionId) external view returns (uint256 usdValue, bool available) {
        Auction storage auction = _auctions[auctionId];
        if (auction.highestBid == 0) {
            return (0, true);
        }
        return quoteUsd(auction.paymentToken, auction.highestBid);
    }

    /// @notice 根据美元金额返回手续费基点（动态手续费）
    function feeBpsForUsd(uint256 usdValue) public pure returns (uint256) {
        if (usdValue < FEE_TIER1_LIMIT_USD) {
            return FEE_LOW_BPS;
        }
        if (usdValue < FEE_TIER2_LIMIT_USD) {
            return FEE_MID_BPS;
        }
        return FEE_HIGH_BPS;
    }

    /// @notice 预览一笔成交价的手续费
    function calculateFee(address token, uint256 amount) external view returns (uint256 fee, uint256 usdValue) {
        return _calculateFee(token, amount);
    }

    // ---------------------------------------------------------------- 管理

    /// @notice 更新价格预言机
    function setPriceOracle(address oracle) external onlyOwner {
        priceOracle = PriceOracle(oracle);
        emit PriceOracleUpdated(oracle);
    }

    /// @notice 更新手续费收款地址
    function setFeeRecipient(address feeRecipient_) external onlyOwner {
        feeRecipient = feeRecipient_;
        emit FeeRecipientUpdated(feeRecipient_);
    }

    // ---------------------------------------------------------------- 内部

    /// @dev 出价核心逻辑，ETH 与 ERC20 共用；资金已在合约内托管
    function _placeBid(uint256 auctionId, Auction storage auction, uint256 amount) internal {
        require(block.timestamp < auction.endTime, "NFTAuction: auction ended");

        uint256 minimum = auction.highestBid == 0 ? auction.startingPrice : auction.highestBid;
        require(amount >= minimum, "NFTAuction: bid below minimum");
        if (auction.highestBid > 0) {
            require(amount > auction.highestBid, "NFTAuction: bid not higher than current");
        }
        _validateBid(auction, amount);

        address previousBidder = auction.highestBidder;
        uint256 previousBid = auction.highestBid;

        auction.highestBid = amount;
        auction.highestBidder = msg.sender;

        if (previousBidder != address(0)) {
            _payout(auction.paymentToken, previousBidder, previousBid);
        }

        (uint256 usdValue,) = quoteUsd(auction.paymentToken, amount);
        emit BidPlaced(auctionId, msg.sender, amount, usdValue);
    }

    /// @dev 出价校验钩子，V2 会覆盖它来强制最小加价比例
    function _validateBid(Auction storage, uint256) internal view virtual {}

    /// @dev 计算手续费：取价失败时回落到最低档，避免预言机异常阻塞成交
    function _calculateFee(address token, uint256 amount) internal view returns (uint256 fee, uint256 usdValue) {
        (uint256 usd, bool available) = quoteUsd(token, amount);
        uint256 bps = available ? feeBpsForUsd(usd) : FEE_LOW_BPS;
        return ((amount * bps) / BPS_DENOMINATOR, usd);
    }

    /// @dev ETH / ERC20 通用转账
    function _payout(address token, address to, uint256 amount) private {
        if (amount == 0) {
            return;
        }
        if (token == address(0)) {
            (bool ok,) = to.call{value: amount}("");
            require(ok, "NFTAuction: ETH transfer failed");
        } else {
            IERC20(token).safeTransfer(to, amount);
        }
    }

    /// @dev 取一个仍在进行中的拍卖
    function _activeAuction(uint256 auctionId) private view returns (Auction storage auction) {
        auction = _auctions[auctionId];
        require(auction.seller != address(0), "NFTAuction: unknown auction");
        require(auction.state == AuctionState.Created, "NFTAuction: auction not active");
        require(block.timestamp < auction.endTime, "NFTAuction: auction ended");
    }

    /// @dev 只有 owner 能升级实现合约（UUPS）
    function _authorizeUpgrade(address newImplementation) internal override onlyOwner {}
}
