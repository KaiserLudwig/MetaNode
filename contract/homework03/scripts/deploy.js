const fs = require("fs");
const path = require("path");
const { ethers, upgrades, network } = require("hardhat");

// 已在链上核实过的 Sepolia Chainlink 喂价（description / decimals / latestRoundData 均正常）
const SEPOLIA_FEEDS = {
  ETH_USD: "0x694AA1769357215DE4FAC081bf1f309aDC325306",
  LINK_USD: "0xc59E3633BAAC79493d908e63626716e204A45EdF",
  USDC_USD: "0xA2F78ab2355fe2f984D808B5CeE7FD0A93D5270E",
  DAI_USD: "0x14866185B1962B63C3Ea9E03Bc1da838bab34C19",
};

async function main() {
  const [deployer] = await ethers.getSigners();
  const chainId = (await ethers.provider.getNetwork()).chainId;
  const balance = await ethers.provider.getBalance(deployer.address);

  console.log("网络:", network.name, "| chainId:", chainId.toString());
  console.log("部署账户:", deployer.address);
  console.log("账户余额:", ethers.formatEther(balance), "ETH");

  // 1) 价格预言机
  const oracle = await ethers.deployContract("PriceOracle", [deployer.address]);
  await oracle.waitForDeployment();
  console.log("PriceOracle:", oracle.target);

  // 2) 配置喂价：Sepolia 用真实 Chainlink feed，本地网络用 mock
  let ethFeedAddress;
  if (network.name === "sepolia") {
    ethFeedAddress = SEPOLIA_FEEDS.ETH_USD;
  } else {
    const mockFeed = await ethers.deployContract("MockV3Aggregator", [8, 2000n * 10n ** 8n]);
    await mockFeed.waitForDeployment();
    ethFeedAddress = mockFeed.target;
    console.log("MockV3Aggregator (ETH/USD = 2000):", ethFeedAddress);
  }
  await (await oracle.setPriceFeed(ethers.ZeroAddress, ethFeedAddress)).wait();
  console.log("已配置 ETH/USD 喂价:", ethFeedAddress);
  console.log("当前 ETH 价格(USD, 8位小数):", (await oracle.getPrice(ethers.ZeroAddress)).toString());

  // 3) NFT 合约
  const nft = await ethers.deployContract("AuctionNFT", [deployer.address]);
  await nft.waitForDeployment();
  console.log("AuctionNFT:", nft.target);

  // 4) 拍卖市场（UUPS 代理）
  const NFTAuction = await ethers.getContractFactory("NFTAuction");
  const auction = await upgrades.deployProxy(
    NFTAuction,
    [deployer.address, oracle.target, deployer.address],
    { kind: "uups" },
  );
  await auction.waitForDeployment();
  const implementationAddress = await upgrades.erc1967.getImplementationAddress(auction.target);
  console.log("NFTAuction 代理(Proxy):", auction.target);
  console.log("NFTAuction 实现(Implementation):", implementationAddress);
  console.log("合约版本:", await auction.version());

  // 5) 可选：部署一个演示用 ERC20，并把它接到 USDC/USD 喂价上
  let demoTokenAddress = "0x0000000000000000000000000000000000000000";
  if (process.env.DEMO_TOKEN === "1") {
    const token = await ethers.deployContract("MockERC20", ["Demo USD", "dUSD", 18]);
    await token.waitForDeployment();
    demoTokenAddress = token.target;

    const tokenFeed = network.name === "sepolia" ? SEPOLIA_FEEDS.USDC_USD : ethFeedAddress;
    await (await oracle.setPriceFeed(demoTokenAddress, tokenFeed)).wait();
    await (await token.mint(deployer.address, ethers.parseEther("1000000"))).wait();
    console.log("演示 ERC20 (dUSD):", demoTokenAddress);
  }

  // 6) 铸造一个演示 NFT
  let demoTokenId = null;
  if (process.env.MINT_DEMO !== "0") {
    const receipt = await (await nft.mint(deployer.address, "ipfs://demo-nft/1")).wait();
    demoTokenId = 1n;
    console.log("已铸造演示 NFT tokenId=1，交易:", receipt.hash);
  }

  // 7) 记录部署结果
  const record = {
    network: network.name,
    chainId: chainId.toString(),
    deployer: deployer.address,
    deployedAt: new Date().toISOString(),
    contracts: {
      priceOracle: oracle.target,
      auctionNFT: nft.target,
      nftAuctionProxy: auction.target,
      nftAuctionImplementation: implementationAddress,
      demoErc20: demoTokenAddress,
    },
    feeds: {
      ethUsd: ethFeedAddress,
      ...(network.name === "sepolia" ? SEPOLIA_FEEDS : {}),
    },
    demoNftTokenId: demoTokenId ? demoTokenId.toString() : null,
  };

  const outDir = path.join(__dirname, "..", "deployments");
  fs.mkdirSync(outDir, { recursive: true });
  const outFile = path.join(outDir, `${network.name}.json`);
  fs.writeFileSync(outFile, JSON.stringify(record, null, 2));
  console.log("部署结果已写入:", outFile);

  if (network.name === "sepolia") {
    console.log("\n区块浏览器：");
    console.log("  NFT        https://sepolia.etherscan.io/address/" + nft.target);
    console.log("  预言机     https://sepolia.etherscan.io/address/" + oracle.target);
    console.log("  拍卖(代理) https://sepolia.etherscan.io/address/" + auction.target);
  }
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
