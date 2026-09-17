const fs = require("fs");
const path = require("path");
const { ethers, upgrades, network } = require("hardhat");

/// UUPS 升级脚本：把已部署的拍卖代理升级到 NFTAuctionV2，并初始化新增的最小加价比例
/// 用法：npx hardhat run scripts/upgrade.js --network sepolia
async function main() {
  const recordFile = path.join(__dirname, "..", "deployments", `${network.name}.json`);
  const proxyAddress = process.env.PROXY || (fs.existsSync(recordFile) ? JSON.parse(fs.readFileSync(recordFile)).contracts.nftAuctionProxy : null);

  if (!proxyAddress) {
    throw new Error("找不到代理地址：请先部署，或通过 PROXY 环境变量指定");
  }

  const incrementBps = BigInt(process.env.MIN_BID_INCREMENT_BPS || "500"); // 默认 5%
  const [deployer] = await ethers.getSigners();

  console.log("网络:", network.name, "| 操作账户:", deployer.address);
  console.log("升级前版本:", await (await ethers.getContractAt("NFTAuction", proxyAddress)).version());

  const NFTAuctionV2 = await ethers.getContractFactory("NFTAuctionV2");
  const upgraded = await upgrades.upgradeProxy(proxyAddress, NFTAuctionV2, { kind: "uups" });
  await upgraded.waitForDeployment();

  await (await upgraded.initializeV2(incrementBps)).wait();

  const implementationAddress = await upgrades.erc1967.getImplementationAddress(proxyAddress);
  console.log("代理地址:", proxyAddress);
  console.log("新实现地址:", implementationAddress);
  console.log("升级后版本:", await upgraded.version());
  console.log("最小加价比例(bps):", (await upgraded.minBidIncrementBps()).toString());
  console.log("拍卖数量(状态保留):", (await upgraded.auctionCount()).toString());

  if (network.name === "sepolia") {
    console.log("区块浏览器: https://sepolia.etherscan.io/address/" + proxyAddress);
  }
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
