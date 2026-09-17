const { expect } = require("chai");
const { ethers } = require("hardhat");
const { loadFixture, time } = require("@nomicfoundation/hardhat-network-helpers");

const ETH_USD = 2000n * 10n ** 8n;
const USD_USD = 1n * 10n ** 8n;

describe("PriceOracle", function () {
  async function deployFixture() {
    const [owner, other] = await ethers.getSigners();
    const oracle = await ethers.deployContract("PriceOracle", [owner.address]);
    const ethFeed = await ethers.deployContract("MockV3Aggregator", [8, ETH_USD]);
    const usdToken = await ethers.deployContract("MockERC20", ["Mock USD", "mUSD", 6]);
    const usdFeed = await ethers.deployContract("MockV3Aggregator", [8, USD_USD]);
    const eth18Feed = await ethers.deployContract("MockV3Aggregator", [18, 2000n * 10n ** 18n]);

    return { oracle, ethFeed, usdFeed, eth18Feed, usdToken, owner, other };
  }

  it("owner 可以设置价格源，其他人不行", async function () {
    const { oracle, ethFeed, other } = await loadFixture(deployFixture);

    await expect(oracle.connect(other).setPriceFeed(ethers.ZeroAddress, ethFeed.target))
      .to.be.revertedWithCustomError(oracle, "OwnableUnauthorizedAccount")
      .withArgs(other.address);

    await expect(oracle.setPriceFeed(ethers.ZeroAddress, ethFeed.target))
      .to.emit(oracle, "PriceFeedUpdated")
      .withArgs(ethers.ZeroAddress, ethFeed.target);

    expect(await oracle.priceFeeds(ethers.ZeroAddress)).to.equal(ethFeed.target);
  });

  it("价格源不能设置为零地址", async function () {
    const { oracle } = await loadFixture(deployFixture);
    await expect(oracle.setPriceFeed(ethers.ZeroAddress, ethers.ZeroAddress)).to.be.revertedWith(
      "PriceOracle: feed is zero",
    );
  });

  it("未设置价格源时读取价格会回滚", async function () {
    const { oracle, usdToken } = await loadFixture(deployFixture);
    await expect(oracle.getPrice(usdToken.target)).to.be.revertedWith("PriceOracle: feed not set");
  });

  it("getPrice 返回 8 位小数的美元价格", async function () {
    const { oracle, ethFeed, usdToken, usdFeed } = await loadFixture(deployFixture);
    await oracle.setPriceFeed(ethers.ZeroAddress, ethFeed.target);
    await oracle.setPriceFeed(usdToken.target, usdFeed.target);

    expect(await oracle.getPrice(ethers.ZeroAddress)).to.equal(ETH_USD);
    expect(await oracle.getPrice(usdToken.target)).to.equal(USD_USD);
  });

  it("价格源精度高于 8 位时会自动缩放", async function () {
    const { oracle, eth18Feed } = await loadFixture(deployFixture);
    await oracle.setPriceFeed(ethers.ZeroAddress, eth18Feed.target);
    expect(await oracle.getPrice(ethers.ZeroAddress)).to.equal(ETH_USD);
  });

  it("convertToUsd 按代币精度换算", async function () {
    const { oracle, ethFeed, usdToken, usdFeed } = await loadFixture(deployFixture);
    await oracle.setPriceFeed(ethers.ZeroAddress, ethFeed.target);
    await oracle.setPriceFeed(usdToken.target, usdFeed.target);

    // 1 ETH = 2000 USD
    expect(await oracle.convertToUsd(ethers.ZeroAddress, ethers.parseEther("1"))).to.equal(2000n * 10n ** 8n);
    // 0.5 ETH = 1000 USD
    expect(await oracle.convertToUsd(ethers.ZeroAddress, ethers.parseEther("0.5"))).to.equal(1000n * 10n ** 8n);
    // mUSD 精度是 6：100 mUSD = 100 USD
    expect(await oracle.convertToUsd(usdToken.target, 100n * 10n ** 6n)).to.equal(100n * 10n ** 8n);
  });

  it("decimalsOf 对 ETH 返回 18，对 ERC20 读取 decimals()", async function () {
    const { oracle, usdToken } = await loadFixture(deployFixture);
    expect(await oracle.decimalsOf(ethers.ZeroAddress)).to.equal(18);
    expect(await oracle.decimalsOf(usdToken.target)).to.equal(6);
  });

  it("价格过期会回滚", async function () {
    const { oracle, ethFeed } = await loadFixture(deployFixture);
    await oracle.setPriceFeed(ethers.ZeroAddress, ethFeed.target);

    await time.increase(2 * 60 * 60); // 超过 MAX_STALENESS（1 小时）
    await expect(oracle.getPrice(ethers.ZeroAddress)).to.be.revertedWith("PriceOracle: stale price");
  });

  it("价格非正数会回滚", async function () {
    const { oracle, ethFeed } = await loadFixture(deployFixture);
    await oracle.setPriceFeed(ethers.ZeroAddress, ethFeed.target);
    await ethFeed.updateAnswer(0);
    await expect(oracle.getPrice(ethers.ZeroAddress)).to.be.revertedWith("PriceOracle: invalid price");
  });
});
