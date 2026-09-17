const { expect } = require("chai");
const { ethers, upgrades } = require("hardhat");
const { loadFixture, time } = require("@nomicfoundation/hardhat-network-helpers");

const ETH_USD = 2000n * 10n ** 8n; // ETH = 2000 USD
const TOKEN_USD = 1n * 10n ** 8n; // 1 mUSD = 1 USD
const HOUR = 3600;

describe("NFTAuction", function () {
  async function deployFixture() {
    const [owner, seller, bidder1, bidder2, feeRecipient, stranger] = await ethers.getSigners();

    const nft = await ethers.deployContract("AuctionNFT", [owner.address]);
    const oracle = await ethers.deployContract("PriceOracle", [owner.address]);
    const ethFeed = await ethers.deployContract("MockV3Aggregator", [8, ETH_USD]);
    const erc20 = await ethers.deployContract("MockERC20", ["Mock USD", "mUSD", 18]);
    const erc20Feed = await ethers.deployContract("MockV3Aggregator", [8, TOKEN_USD]);

    await oracle.setPriceFeed(ethers.ZeroAddress, ethFeed.target);
    await oracle.setPriceFeed(erc20.target, erc20Feed.target);

    const NFTAuction = await ethers.getContractFactory("NFTAuction");
    const auction = await upgrades.deployProxy(
      NFTAuction,
      [owner.address, oracle.target, feeRecipient.address],
      { kind: "uups" },
    );
    await auction.waitForDeployment();

    await nft.mint(seller.address, "ipfs://token/1");
    await nft.connect(seller).setApprovalForAll(auction.target, true);

    await erc20.mint(bidder1.address, ethers.parseEther("1000"));
    await erc20.mint(bidder2.address, ethers.parseEther("1000"));
    await erc20.connect(bidder1).approve(auction.target, ethers.MaxUint256);
    await erc20.connect(bidder2).approve(auction.target, ethers.MaxUint256);

    return { nft, oracle, ethFeed, erc20, erc20Feed, auction, owner, seller, bidder1, bidder2, feeRecipient, stranger };
  }

  async function createEthAuction(f, startingPrice = ethers.parseEther("0.1"), duration = HOUR) {
    await f.auction
      .connect(f.seller)
      .createAuction(f.nft.target, 1n, ethers.ZeroAddress, startingPrice, duration);
    return 1n;
  }

  async function createErc20Auction(f, startingPrice = ethers.parseEther("100"), duration = HOUR) {
    await f.auction.connect(f.seller).createAuction(f.nft.target, 1n, f.erc20.target, startingPrice, duration);
    return 1n;
  }

  describe("初始化", function () {
    it("部署后 owner / 预言机 / 收款地址正确，版本为 1.0.0", async function () {
      const { auction, owner, oracle, feeRecipient } = await loadFixture(deployFixture);
      expect(await auction.owner()).to.equal(owner.address);
      expect(await auction.priceOracle()).to.equal(oracle.target);
      expect(await auction.feeRecipient()).to.equal(feeRecipient.address);
      expect(await auction.auctionCount()).to.equal(0n);
      expect(await auction.version()).to.equal("1.0.0");
    });

    it("initialize 不能被调用两次", async function () {
      const { auction, owner, oracle, feeRecipient } = await loadFixture(deployFixture);
      await expect(
        auction.initialize(owner.address, oracle.target, feeRecipient.address),
      ).to.be.revertedWithCustomError(auction, "InvalidInitialization");
    });

    it("实现合约本身禁止初始化（_disableInitializers）", async function () {
      const NFTAuction = await ethers.getContractFactory("NFTAuction");
      const implementation = await NFTAuction.deploy();
      const [, , , , , owner] = await ethers.getSigners();
      await expect(
        implementation.initialize(owner.address, ethers.ZeroAddress, owner.address),
      ).to.be.revertedWithCustomError(implementation, "InvalidInitialization");
    });
  });

  describe("创建拍卖", function () {
    it("成功创建：NFT 转入托管，字段与事件正确", async function () {
      const f = await loadFixture(deployFixture);
      const { auction, nft, seller } = f;
      const startingPrice = ethers.parseEther("0.5");

      const tx = auction.connect(seller).createAuction(nft.target, 1n, ethers.ZeroAddress, startingPrice, HOUR);
      await expect(tx).to.emit(auction, "AuctionCreated");
      await tx;

      expect(await nft.ownerOf(1n)).to.equal(auction.target);

      const info = await auction.getAuction(1n);
      expect(info.seller).to.equal(seller.address);
      expect(info.nft).to.equal(nft.target);
      expect(info.tokenId).to.equal(1n);
      expect(info.paymentToken).to.equal(ethers.ZeroAddress);
      expect(info.startingPrice).to.equal(startingPrice);
      expect(info.state).to.equal(0n); // Created
      expect(await auction.auctionCount()).to.equal(1n);
    });

    it("非 NFT 持有者不能上架", async function () {
      const f = await loadFixture(deployFixture);
      await expect(
        f.auction.connect(f.stranger).createAuction(f.nft.target, 1n, ethers.ZeroAddress, 1n, HOUR),
      ).to.be.revertedWith("NFTAuction: not the token owner");
    });

    it("未授权拍卖合约时不能上架", async function () {
      const f = await loadFixture(deployFixture);
      await f.nft.connect(f.seller).setApprovalForAll(f.auction.target, false);
      await expect(
        f.auction.connect(f.seller).createAuction(f.nft.target, 1n, ethers.ZeroAddress, 1n, HOUR),
      ).to.be.revertedWith("NFTAuction: not approved");
    });

    it("拍卖时长超出范围会回滚", async function () {
      const f = await loadFixture(deployFixture);
      await expect(
        f.auction.connect(f.seller).createAuction(f.nft.target, 1n, ethers.ZeroAddress, 1n, 30),
      ).to.be.revertedWith("NFTAuction: invalid duration");

      await expect(
        f.auction.connect(f.seller).createAuction(f.nft.target, 1n, ethers.ZeroAddress, 1n, 31 * 24 * HOUR),
      ).to.be.revertedWith("NFTAuction: invalid duration");
    });

    it("起拍价为 0 会回滚", async function () {
      const f = await loadFixture(deployFixture);
      await expect(
        f.auction.connect(f.seller).createAuction(f.nft.target, 1n, ethers.ZeroAddress, 0, HOUR),
      ).to.be.revertedWith("NFTAuction: invalid starting price");
    });

    it("不存在的 tokenId 会回滚", async function () {
      const f = await loadFixture(deployFixture);
      await expect(
        f.auction.connect(f.seller).createAuction(f.nft.target, 999n, ethers.ZeroAddress, 1n, HOUR),
      ).to.be.revertedWithCustomError(f.nft, "ERC721NonexistentToken");
    });
  });

  describe("ETH 出价", function () {
    it("低于起拍价会回滚", async function () {
      const f = await loadFixture(deployFixture);
      const id = await createEthAuction(f, ethers.parseEther("0.1"));
      await expect(
        f.auction.connect(f.bidder1).bidEth(id, { value: ethers.parseEther("0.05") }),
      ).to.be.revertedWith("NFTAuction: bid below minimum");
    });

    it("等于起拍价可以成交，并记录美元估值", async function () {
      const f = await loadFixture(deployFixture);
      const id = await createEthAuction(f, ethers.parseEther("0.1"));

      await expect(f.auction.connect(f.bidder1).bidEth(id, { value: ethers.parseEther("0.1") }))
        .to.emit(f.auction, "BidPlaced")
        .withArgs(id, f.bidder1.address, ethers.parseEther("0.1"), 200n * 10n ** 8n);

      const info = await f.auction.getAuction(id);
      expect(info.highestBidder).to.equal(f.bidder1.address);
      expect(info.highestBid).to.equal(ethers.parseEther("0.1"));
      expect(await ethers.provider.getBalance(f.auction.target)).to.equal(ethers.parseEther("0.1"));
    });

    it("被超越的出价会立即退回", async function () {
      const f = await loadFixture(deployFixture);
      const id = await createEthAuction(f);

      await f.auction.connect(f.bidder1).bidEth(id, { value: ethers.parseEther("0.1") });
      const before = await ethers.provider.getBalance(f.bidder1.address);

      await f.auction.connect(f.bidder2).bidEth(id, { value: ethers.parseEther("0.2") });
      const after = await ethers.provider.getBalance(f.bidder1.address);

      expect(after - before).to.equal(ethers.parseEther("0.1"));
    });

    it("出价不高于当前最高价会回滚", async function () {
      const f = await loadFixture(deployFixture);
      const id = await createEthAuction(f);
      await f.auction.connect(f.bidder1).bidEth(id, { value: ethers.parseEther("0.1") });

      await expect(
        f.auction.connect(f.bidder2).bidEth(id, { value: ethers.parseEther("0.1") }),
      ).to.be.revertedWith("NFTAuction: bid not higher than current");
    });

    it("拍卖结束后不能出价", async function () {
      const f = await loadFixture(deployFixture);
      const id = await createEthAuction(f);
      await time.increase(HOUR + 1);

      await expect(
        f.auction.connect(f.bidder1).bidEth(id, { value: ethers.parseEther("0.2") }),
      ).to.be.revertedWith("NFTAuction: auction ended");
    });

    it("未知拍卖会回滚", async function () {
      const f = await loadFixture(deployFixture);
      await expect(
        f.auction.connect(f.bidder1).bidEth(99n, { value: 1n }),
      ).to.be.revertedWith("NFTAuction: unknown auction");
    });

    it("ERC20 拍卖不能用 ETH 出价", async function () {
      const f = await loadFixture(deployFixture);
      const id = await createErc20Auction(f);
      await expect(
        f.auction.connect(f.bidder1).bidEth(id, { value: ethers.parseEther("1") }),
      ).to.be.revertedWith("NFTAuction: not an ETH auction");
    });
  });

  describe("ERC20 出价", function () {
    it("出价成功后代币由合约托管", async function () {
      const f = await loadFixture(deployFixture);
      const id = await createErc20Auction(f, ethers.parseEther("100"));

      await f.auction.connect(f.bidder1).bidErc20(id, ethers.parseEther("100"));
      expect(await f.erc20.balanceOf(f.auction.target)).to.equal(ethers.parseEther("100"));

      const info = await f.auction.getAuction(id);
      expect(info.highestBid).to.equal(ethers.parseEther("100"));
      expect(info.highestBidder).to.equal(f.bidder1.address);
    });

    it("低于起拍价会回滚", async function () {
      const f = await loadFixture(deployFixture);
      const id = await createErc20Auction(f, ethers.parseEther("100"));
      await expect(
        f.auction.connect(f.bidder1).bidErc20(id, ethers.parseEther("99")),
      ).to.be.revertedWith("NFTAuction: bid below minimum");
    });

    it("被超越的出价会退回 ERC20", async function () {
      const f = await loadFixture(deployFixture);
      const id = await createErc20Auction(f, ethers.parseEther("100"));

      await f.auction.connect(f.bidder1).bidErc20(id, ethers.parseEther("100"));
      await f.auction.connect(f.bidder2).bidErc20(id, ethers.parseEther("150"));

      expect(await f.erc20.balanceOf(f.bidder1.address)).to.equal(ethers.parseEther("1000"));
      expect(await f.erc20.balanceOf(f.auction.target)).to.equal(ethers.parseEther("150"));
    });

    it("没有授权或余额不足会回滚", async function () {
      const f = await loadFixture(deployFixture);
      const id = await createErc20Auction(f, ethers.parseEther("100"));
      await expect(
        f.auction.connect(f.stranger).bidErc20(id, ethers.parseEther("100")),
      ).to.be.revertedWithCustomError(f.erc20, "ERC20InsufficientAllowance");
    });

    it("ETH 拍卖不能用 ERC20 出价", async function () {
      const f = await loadFixture(deployFixture);
      const id = await createEthAuction(f);
      await expect(
        f.auction.connect(f.bidder1).bidErc20(id, ethers.parseEther("100")),
      ).to.be.revertedWith("NFTAuction: not an ERC20 auction");
    });
  });

  describe("结束拍卖", function () {
    it("未到结束时间不能结算", async function () {
      const f = await loadFixture(deployFixture);
      const id = await createEthAuction(f);
      await expect(f.auction.endAuction(id)).to.be.revertedWith("NFTAuction: auction still running");
    });

    it("无人出价时 NFT 退回卖家", async function () {
      const f = await loadFixture(deployFixture);
      const id = await createEthAuction(f);
      await time.increase(HOUR + 1);

      await expect(f.auction.connect(f.stranger).endAuction(id))
        .to.emit(f.auction, "AuctionEnded")
        .withArgs(id, ethers.ZeroAddress, 0n, 0n);

      expect(await f.nft.ownerOf(1n)).to.equal(f.seller.address);
      expect((await f.auction.getAuction(id)).state).to.equal(1n); // Ended
    });

    it("ETH 拍卖结算：NFT 给赢家，卖家收到扣除手续费后的资金", async function () {
      const f = await loadFixture(deployFixture);
      const id = await createEthAuction(f, ethers.parseEther("0.1"));
      await f.auction.connect(f.bidder1).bidEth(id, { value: ethers.parseEther("0.1") });
      await time.increase(HOUR + 1);
      await f.ethFeed.updateAnswer(ETH_USD); // 刷新价格，模拟 Chainlink 持续喂价

      // 0.1 ETH = 200 USD → 1% 手续费
      const sellerBefore = await ethers.provider.getBalance(f.seller.address);
      const feeBefore = await ethers.provider.getBalance(f.feeRecipient.address);

      await f.auction.connect(f.stranger).endAuction(id);

      const fee = (ethers.parseEther("0.1") * 100n) / 10000n;
      const proceeds = ethers.parseEther("0.1") - fee;

      expect(await f.nft.ownerOf(1n)).to.equal(f.bidder1.address);
      expect((await ethers.provider.getBalance(f.seller.address)) - sellerBefore).to.equal(proceeds);
      expect((await ethers.provider.getBalance(f.feeRecipient.address)) - feeBefore).to.equal(fee);
      expect(await ethers.provider.getBalance(f.auction.target)).to.equal(0n);
    });

    it("ERC20 拍卖结算：代币按比例分给卖家与平台", async function () {
      const f = await loadFixture(deployFixture);
      const id = await createErc20Auction(f, ethers.parseEther("99"));
      await f.auction.connect(f.bidder1).bidErc20(id, ethers.parseEther("99"));
      await time.increase(HOUR + 1);
      await f.erc20Feed.updateAnswer(TOKEN_USD); // 刷新价格

      // 99 mUSD = 99 USD → 0.5% 手续费
      await f.auction.connect(f.stranger).endAuction(id);

      const fee = (ethers.parseEther("99") * 50n) / 10000n;
      expect(await f.erc20.balanceOf(f.feeRecipient.address)).to.equal(fee);
      expect(await f.erc20.balanceOf(f.seller.address)).to.equal(ethers.parseEther("99") - fee);
      expect(await f.erc20.balanceOf(f.auction.target)).to.equal(0n);
      expect(await f.nft.ownerOf(1n)).to.equal(f.bidder1.address);
    });

    it("已结束的拍卖不能重复结算", async function () {
      const f = await loadFixture(deployFixture);
      const id = await createEthAuction(f);
      await f.auction.connect(f.bidder1).bidEth(id, { value: ethers.parseEther("0.1") });
      await time.increase(HOUR + 1);
      await f.auction.endAuction(id);

      await expect(f.auction.endAuction(id)).to.be.revertedWith("NFTAuction: auction not active");
    });
  });

  describe("动态手续费（额外挑战）", function () {
    it("按美元金额分三档", async function () {
      const { auction } = await loadFixture(deployFixture);
      expect(await auction.feeBpsForUsd(50n * 10n ** 8n)).to.equal(50n); // < 100 USD → 0.5%
      expect(await auction.feeBpsForUsd(500n * 10n ** 8n)).to.equal(100n); // < 1000 USD → 1%
      expect(await auction.feeBpsForUsd(5000n * 10n ** 8n)).to.equal(200n); // >= 1000 USD → 2%
    });

    it("calculateFee 与实际结算一致（ETH）", async function () {
      const f = await loadFixture(deployFixture);
      const cases = [
        { amount: ethers.parseEther("0.01"), expectedBps: 50n }, // 20 USD
        { amount: ethers.parseEther("0.1"), expectedBps: 100n }, // 200 USD
        { amount: ethers.parseEther("1"), expectedBps: 200n }, // 2000 USD
      ];

      for (const { amount, expectedBps } of cases) {
        const [fee, usdValue] = await f.auction.calculateFee(ethers.ZeroAddress, amount);
        expect(usdValue).to.equal((amount * ETH_USD) / ethers.parseEther("1"));
        expect(fee).to.equal((amount * expectedBps) / 10000n);
      }
    });

    it("取价失败时回落到最低档手续费，不阻塞结算", async function () {
      const f = await loadFixture(deployFixture);
      // 用一个没有配置 feed 的 ERC20 拍卖
      const unknownToken = await ethers.deployContract("MockERC20", ["Unknown", "UNK", 18]);
      await unknownToken.mint(f.bidder1.address, ethers.parseEther("10"));
      await unknownToken.connect(f.bidder1).approve(f.auction.target, ethers.MaxUint256);

      await f.auction
        .connect(f.seller)
        .createAuction(f.nft.target, 1n, unknownToken.target, ethers.parseEther("1"), HOUR);
      const id = 1n;

      const [usdValue, available] = await f.auction.quoteUsd(unknownToken.target, ethers.parseEther("1"));
      expect(available).to.equal(false);
      expect(usdValue).to.equal(0n);

      await f.auction.connect(f.bidder1).bidErc20(id, ethers.parseEther("1"));
      await time.increase(HOUR + 1);
      await f.auction.endAuction(id);

      const fee = (ethers.parseEther("1") * 50n) / 10000n; // 回落 0.5%
      expect(await unknownToken.balanceOf(f.seller.address)).to.equal(ethers.parseEther("1") - fee);
      expect(await f.nft.ownerOf(1n)).to.equal(f.bidder1.address);
    });

    it("价格过期时同样回落到最低档手续费", async function () {
      const f = await loadFixture(deployFixture);
      const amount = ethers.parseEther("0.1"); // 若价格新鲜应为 200 USD → 1%

      let [fee] = await f.auction.calculateFee(ethers.ZeroAddress, amount);
      expect(fee).to.equal((amount * 100n) / 10000n);

      await time.increase(2 * HOUR); // 超过预言机 1 小时的过期阈值，且不刷新价格
      const [, available] = await f.auction.quoteUsd(ethers.ZeroAddress, amount);
      expect(available).to.equal(false);

      [fee] = await f.auction.calculateFee(ethers.ZeroAddress, amount);
      expect(fee).to.equal((amount * 50n) / 10000n);
    });

    it("收款地址为零时免收手续费", async function () {
      const f = await loadFixture(deployFixture);
      const id = await createEthAuction(f, ethers.parseEther("0.1"));
      await f.auction.connect(f.bidder1).bidEth(id, { value: ethers.parseEther("0.1") });
      await f.auction.setFeeRecipient(ethers.ZeroAddress);
      await time.increase(HOUR + 1);

      const sellerBefore = await ethers.provider.getBalance(f.seller.address);
      await f.auction.endAuction(id);
      expect((await ethers.provider.getBalance(f.seller.address)) - sellerBefore).to.equal(
        ethers.parseEther("0.1"),
      );
    });
  });

  describe("取消拍卖", function () {
    it("卖家可以在无人出价时取消，NFT 退回", async function () {
      const f = await loadFixture(deployFixture);
      const id = await createEthAuction(f);

      await expect(f.auction.connect(f.seller).cancelAuction(id))
        .to.emit(f.auction, "AuctionCancelled")
        .withArgs(id);

      expect(await f.nft.ownerOf(1n)).to.equal(f.seller.address);
      expect((await f.auction.getAuction(id)).state).to.equal(2n); // Cancelled
    });

    it("非卖家不能取消", async function () {
      const f = await loadFixture(deployFixture);
      const id = await createEthAuction(f);
      await expect(f.auction.connect(f.stranger).cancelAuction(id)).to.be.revertedWith(
        "NFTAuction: not allowed",
      );
    });

    it("平台 owner 可以取消无人出价的拍卖", async function () {
      const f = await loadFixture(deployFixture);
      const id = await createEthAuction(f);
      await f.auction.connect(f.owner).cancelAuction(id);
      expect(await f.nft.ownerOf(1n)).to.equal(f.seller.address);
    });

    it("已有出价时不能取消", async function () {
      const f = await loadFixture(deployFixture);
      const id = await createEthAuction(f);
      await f.auction.connect(f.bidder1).bidEth(id, { value: ethers.parseEther("0.1") });

      await expect(f.auction.connect(f.seller).cancelAuction(id)).to.be.revertedWith(
        "NFTAuction: already has bids",
      );
    });
  });

  describe("价格换算与查询", function () {
    it("quoteUsd 换算 ETH 与 ERC20", async function () {
      const f = await loadFixture(deployFixture);
      const [ethUsd, ethOk] = await f.auction.quoteUsd(ethers.ZeroAddress, ethers.parseEther("1"));
      expect(ethOk).to.equal(true);
      expect(ethUsd).to.equal(2000n * 10n ** 8n);

      const [tokenUsd, tokenOk] = await f.auction.quoteUsd(f.erc20.target, ethers.parseEther("250"));
      expect(tokenOk).to.equal(true);
      expect(tokenUsd).to.equal(250n * 10n ** 8n);
    });

    it("highestBidUsd 返回当前最高出价的美元估值", async function () {
      const f = await loadFixture(deployFixture);
      const id = await createEthAuction(f);

      let [usd, ok] = await f.auction.highestBidUsd(id);
      expect(ok).to.equal(true);
      expect(usd).to.equal(0n);

      await f.auction.connect(f.bidder1).bidEth(id, { value: ethers.parseEther("0.5") });
      [usd, ok] = await f.auction.highestBidUsd(id);
      expect(ok).to.equal(true);
      expect(usd).to.equal(1000n * 10n ** 8n);
    });

    it("预言机价格变化后，同样的出价会落入不同手续费档", async function () {
      const f = await loadFixture(deployFixture);
      const amount = ethers.parseEther("0.1");

      let [fee] = await f.auction.calculateFee(ethers.ZeroAddress, amount);
      expect(fee).to.equal((amount * 100n) / 10000n); // 200 USD → 1%

      await f.ethFeed.updateAnswer(10000n * 10n ** 8n); // ETH 涨到 10000 USD
      [fee] = await f.auction.calculateFee(ethers.ZeroAddress, amount);
      expect(fee).to.equal((amount * 200n) / 10000n); // 1000 USD → 2%
    });
  });

  describe("平台管理", function () {
    it("owner 可以更新预言机与收款地址", async function () {
      const f = await loadFixture(deployFixture);
      const newOracle = await ethers.deployContract("PriceOracle", [f.owner.address]);

      await expect(f.auction.setPriceOracle(newOracle.target))
        .to.emit(f.auction, "PriceOracleUpdated")
        .withArgs(newOracle.target);
      expect(await f.auction.priceOracle()).to.equal(newOracle.target);

      await expect(f.auction.setFeeRecipient(f.stranger.address))
        .to.emit(f.auction, "FeeRecipientUpdated")
        .withArgs(f.stranger.address);
      expect(await f.auction.feeRecipient()).to.equal(f.stranger.address);
    });

    it("非 owner 不能修改平台配置", async function () {
      const f = await loadFixture(deployFixture);
      await expect(f.auction.connect(f.stranger).setPriceOracle(f.stranger.address))
        .to.be.revertedWithCustomError(f.auction, "OwnableUnauthorizedAccount")
        .withArgs(f.stranger.address);
      await expect(f.auction.connect(f.stranger).setFeeRecipient(f.stranger.address))
        .to.be.revertedWithCustomError(f.auction, "OwnableUnauthorizedAccount")
        .withArgs(f.stranger.address);
    });
  });

  describe("UUPS 升级", function () {
    it("升级到 V2：版本变化且状态保留", async function () {
      const f = await loadFixture(deployFixture);
      const id = await createEthAuction(f);
      await f.auction.connect(f.bidder1).bidEth(id, { value: ethers.parseEther("0.1") });

      const NFTAuctionV2 = await ethers.getContractFactory("NFTAuctionV2");
      const upgraded = await upgrades.upgradeProxy(f.auction.target, NFTAuctionV2);
      await upgraded.waitForDeployment();

      expect(await upgraded.version()).to.equal("2.0.0");
      expect(await upgraded.auctionCount()).to.equal(1n);
      const info = await upgraded.getAuction(id);
      expect(info.highestBidder).to.equal(f.bidder1.address);
      expect(info.highestBid).to.equal(ethers.parseEther("0.1"));
      expect(await upgraded.owner()).to.equal(f.owner.address);
      expect(await upgraded.feeRecipient()).to.equal(f.feeRecipient.address);
    });

    it("升级后新增的最小加价规则生效", async function () {
      const f = await loadFixture(deployFixture);
      const id = await createEthAuction(f, ethers.parseEther("0.1"));
      await f.auction.connect(f.bidder1).bidEth(id, { value: ethers.parseEther("0.1") });

      const NFTAuctionV2 = await ethers.getContractFactory("NFTAuctionV2");
      const upgraded = await upgrades.upgradeProxy(f.auction.target, NFTAuctionV2);

      await upgraded.initializeV2(1000n); // 至少加价 10%
      expect(await upgraded.minBidIncrementBps()).to.equal(1000n);

      await expect(
        upgraded.connect(f.bidder2).bidEth(id, { value: ethers.parseEther("0.105") }),
      ).to.be.revertedWith("NFTAuctionV2: bid increment too small");

      await expect(upgraded.connect(f.bidder2).bidEth(id, { value: ethers.parseEther("0.11") })).to.emit(
        upgraded,
        "BidPlaced",
      );
    });

    it("V2 的 initializeV2 只能执行一次", async function () {
      const f = await loadFixture(deployFixture);
      const NFTAuctionV2 = await ethers.getContractFactory("NFTAuctionV2");
      const upgraded = await upgrades.upgradeProxy(f.auction.target, NFTAuctionV2);

      await upgraded.initializeV2(500n);
      await expect(upgraded.initializeV2(600n)).to.be.revertedWithCustomError(
        upgraded,
        "InvalidInitialization",
      );
    });

    it("非 owner 不能升级实现合约", async function () {
      const f = await loadFixture(deployFixture);
      const NFTAuctionV2 = await ethers.getContractFactory("NFTAuctionV2");
      const newImplementation = await NFTAuctionV2.deploy();

      await expect(
        f.auction.connect(f.stranger).upgradeToAndCall(newImplementation.target, "0x"),
      )
        .to.be.revertedWithCustomError(f.auction, "OwnableUnauthorizedAccount")
        .withArgs(f.stranger.address);
    });

    it("V2 最小加价比例仅 owner 可调整", async function () {
      const f = await loadFixture(deployFixture);
      const NFTAuctionV2 = await ethers.getContractFactory("NFTAuctionV2");
      const upgraded = await upgrades.upgradeProxy(f.auction.target, NFTAuctionV2);
      await upgraded.initializeV2(0n);

      await expect(upgraded.connect(f.stranger).setMinBidIncrement(100n))
        .to.be.revertedWithCustomError(upgraded, "OwnableUnauthorizedAccount")
        .withArgs(f.stranger.address);

      await expect(upgraded.setMinBidIncrement(100n))
        .to.emit(upgraded, "MinBidIncrementUpdated")
        .withArgs(100n);

      await expect(upgraded.setMinBidIncrement(6000n)).to.be.revertedWith(
        "NFTAuctionV2: increment too large",
      );
    });
  });
});
