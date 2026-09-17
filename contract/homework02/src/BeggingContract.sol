// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title BeggingContract 讨饭合约
/// @notice 任何人都能向合约捐赠 ETH，合约记录每位捐赠者的累计金额，由 owner 提取全部资金
contract BeggingContract {
    /// @notice 合约所有者（部署者），资金只能由它提取
    address payable public immutable owner;

    /// @notice 捐赠者地址 => 累计捐赠金额（wei）
    mapping(address => uint256) private _donations;

    /// @notice 捐赠者首次捐赠的顺序，用于排行榜
    address[] private _donors;

    /// @notice 该地址是否已经捐赠过
    mapping(address => bool) private _hasDonated;

    /// @notice 累计收到的捐赠总额（wei）
    uint256 public totalDonated;

    /// @notice 捐赠开放时间窗的起点，0 表示不限制
    uint256 public startTime;

    /// @notice 捐赠开放时间窗的终点，0 表示不限制
    uint256 public endTime;

    /// @param donor 捐赠者
    /// @param amount 本次捐赠金额
    /// @param donorTotal 该捐赠者累计金额
    /// @param totalDonated 合约累计金额
    event Donation(address indexed donor, uint256 amount, uint256 donorTotal, uint256 totalDonated);

    /// @param owner 收款人
    /// @param amount 提取金额
    event Withdrawal(address indexed owner, uint256 amount);

    /// @param startTime 新的开放起点
    /// @param endTime 新的开放终点
    event DonationWindowUpdated(uint256 startTime, uint256 endTime);

    /// @notice 仅限合约所有者
    modifier onlyOwner() {
        require(msg.sender == owner, "BeggingContract: caller is not the owner");
        _;
    }

    constructor() {
        owner = payable(msg.sender);
    }

    /// @notice 向合约捐赠 ETH，金额记在调用者名下
    function donate() external payable {
        _donate(msg.sender, msg.value);
    }

    /// @notice 直接把 ETH 转到合约地址同样计为捐赠
    receive() external payable {
        _donate(msg.sender, msg.value);
    }

    /// @notice 所有者提取合约里的全部余额
    /// @dev 按作业要求使用 address.transfer（2300 gas）；生产环境更推荐 call 并处理返回值
    function withdraw() external onlyOwner {
        uint256 balance = address(this).balance;
        require(balance > 0, "BeggingContract: nothing to withdraw");

        emit Withdrawal(owner, balance);
        owner.transfer(balance);
    }

    /// @notice 查询某个地址的累计捐赠金额
    function getDonation(address donor) external view returns (uint256) {
        return _donations[donor];
    }

    /// @notice 当前是否处于可捐赠状态
    /// @dev 时间窗是作业要求的功能，这里的区块时间比较是预期行为
    function isDonationOpen() public view returns (bool) {
        // forge-lint: disable-next-line(block-timestamp)
        if (startTime != 0 && block.timestamp < startTime) {
            return false;
        }
        // forge-lint: disable-next-line(block-timestamp)
        if (endTime != 0 && block.timestamp > endTime) {
            return false;
        }
        return true;
    }

    /// @notice 捐赠者数量
    function donorCount() external view returns (uint256) {
        return _donors.length;
    }

    /// @notice 返回捐赠金额最高的前 3 名（不足 3 人时其余位置为 0）
    function topDonors() external view returns (address[3] memory topAddresses, uint256[3] memory topAmounts) {
        uint256[3] memory best;
        address[3] memory who;

        for (uint256 i = 0; i < _donors.length; i++) {
            address donor = _donors[i];
            uint256 amount = _donations[donor];

            if (amount > best[0]) {
                best[2] = best[1];
                who[2] = who[1];
                best[1] = best[0];
                who[1] = who[0];
                best[0] = amount;
                who[0] = donor;
            } else if (amount > best[1]) {
                best[2] = best[1];
                who[2] = who[1];
                best[1] = amount;
                who[1] = donor;
            } else if (amount > best[2]) {
                best[2] = amount;
                who[2] = donor;
            }
        }

        return (who, best);
    }

    /// @notice 设置捐赠时间窗（可选挑战 3）；传 0 表示该侧不限制
    function setDonationWindow(uint256 newStartTime, uint256 newEndTime) external onlyOwner {
        require(newEndTime == 0 || newStartTime == 0 || newEndTime > newStartTime, "BeggingContract: invalid window");
        startTime = newStartTime;
        endTime = newEndTime;
        emit DonationWindowUpdated(newStartTime, newEndTime);
    }

    /// @dev 记账逻辑：donate() 与 receive() 共用
    function _donate(address donor, uint256 amount) private {
        require(amount > 0, "BeggingContract: donation must be greater than 0");
        require(isDonationOpen(), "BeggingContract: donation window is closed");

        if (!_hasDonated[donor]) {
            _hasDonated[donor] = true;
            _donors.push(donor);
        }

        _donations[donor] += amount;
        totalDonated += amount;

        emit Donation(donor, amount, _donations[donor], totalDonated);
    }
}
