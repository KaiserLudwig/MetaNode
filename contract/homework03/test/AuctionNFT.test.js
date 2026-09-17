const { expect } = require("chai");
const { ethers } = require("hardhat");
const { loadFixture } = require("@nomicfoundation/hardhat-network-helpers");

describe("AuctionNFT", function () {
  async function deployFixture() {
    const [owner, seller, buyer] = await ethers.getSigners();
    const nft = await ethers.deployContract("AuctionNFT", [owner.address]);
    return { nft, owner, seller, buyer };
  }

  it("部署后名称与符号正确，且初始未铸造任何 NFT", async function () {
    const { nft } = await loadFixture(deployFixture);
    expect(await nft.name()).to.equal("MetaNode Auction NFT");
    expect(await nft.symbol()).to.equal("MNAN");
    expect(await nft.totalMinted()).to.equal(0n);
  });

  it("owner 可以铸造 NFT 并设置 tokenURI", async function () {
    const { nft, owner, seller } = await loadFixture(deployFixture);

    await expect(nft.connect(owner).mint(seller.address, "ipfs://token/1"))
      .to.emit(nft, "NFTMinted")
      .withArgs(seller.address, 1n, "ipfs://token/1");

    expect(await nft.ownerOf(1n)).to.equal(seller.address);
    expect(await nft.tokenURI(1n)).to.equal("ipfs://token/1");
    expect(await nft.totalMinted()).to.equal(1n);
    expect(await nft.balanceOf(seller.address)).to.equal(1n);
  });

  it("非 owner 铸造会被拒绝", async function () {
    const { nft, seller } = await loadFixture(deployFixture);
    await expect(nft.connect(seller).mint(seller.address, "ipfs://token/1"))
      .to.be.revertedWithCustomError(nft, "OwnableUnauthorizedAccount")
      .withArgs(seller.address);
  });

  it("NFT 支持转移与授权", async function () {
    const { nft, owner, seller, buyer } = await loadFixture(deployFixture);
    await nft.connect(owner).mint(seller.address, "ipfs://token/1");

    await expect(nft.connect(seller).transferFrom(seller.address, buyer.address, 1n))
      .to.emit(nft, "Transfer")
      .withArgs(seller.address, buyer.address, 1n);
    expect(await nft.ownerOf(1n)).to.equal(buyer.address);

    await nft.connect(buyer).approve(seller.address, 1n);
    expect(await nft.getApproved(1n)).to.equal(seller.address);
    await nft.connect(seller).transferFrom(buyer.address, seller.address, 1n);
    expect(await nft.ownerOf(1n)).to.equal(seller.address);
  });

  it("连续铸造的 tokenId 自增", async function () {
    const { nft, owner, seller } = await loadFixture(deployFixture);
    await nft.connect(owner).mint(seller.address, "ipfs://token/1");
    await nft.connect(owner).mint(seller.address, "ipfs://token/2");
    expect(await nft.totalMinted()).to.equal(2n);
    expect(await nft.tokenURI(2n)).to.equal("ipfs://token/2");
  });

  it("支持 ERC721 与元数据接口", async function () {
    const { nft } = await loadFixture(deployFixture);
    expect(await nft.supportsInterface("0x80ac58cd")).to.equal(true); // ERC721
    expect(await nft.supportsInterface("0x5b5e139f")).to.equal(true); // ERC721Metadata
  });
});
