// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {ERC721} from "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import {ERC721URIStorage} from "@openzeppelin/contracts/token/ERC721/extensions/ERC721URIStorage.sol";
import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";

/// @title AuctionNFT 拍卖市场使用的 ERC721
/// @notice 作业要求：使用 ERC721 标准实现 NFT 合约，支持铸造与转移
contract AuctionNFT is ERC721, ERC721URIStorage, Ownable {
    uint256 private _nextTokenId;

    /// @param to 接收者
    /// @param tokenId 代币编号
    /// @param uri 元数据地址
    event NFTMinted(address indexed to, uint256 indexed tokenId, string uri);

    constructor(address initialOwner) ERC721("MetaNode Auction NFT", "MNAN") Ownable(initialOwner) {}

    /// @notice 铸造一个 NFT（仅 owner 可调用，演示用）
    function mint(address to, string calldata uri) external onlyOwner returns (uint256 tokenId) {
        tokenId = ++_nextTokenId;
        _safeMint(to, tokenId);
        _setTokenURI(tokenId, uri);
        emit NFTMinted(to, tokenId, uri);
    }

    /// @notice 已铸造总量
    function totalMinted() external view returns (uint256) {
        return _nextTokenId;
    }

    function tokenURI(uint256 tokenId) public view override(ERC721, ERC721URIStorage) returns (string memory) {
        return super.tokenURI(tokenId);
    }

    function supportsInterface(bytes4 interfaceId)
        public
        view
        override(ERC721, ERC721URIStorage)
        returns (bool)
    {
        return super.supportsInterface(interfaceId);
    }
}
