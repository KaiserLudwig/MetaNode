// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title Counter 计数器合约
/// @notice 作业 05 任务 2 用：演示用 abigen 生成 Go 绑定后与真实链上合约交互
contract Counter {
    /// @notice 当前计数值
    uint256 private _count;

    /// @notice 部署者，拥有重置权限
    address public immutable owner;

    /// @param by 触发变化的调用者
    /// @param newValue 变化后的计数值
    event CountChanged(address indexed by, uint256 newValue);

    constructor(uint256 initialValue) {
        _count = initialValue;
        owner = msg.sender;
        emit CountChanged(msg.sender, initialValue);
    }

    /// @notice 计数加一
    function inc() external {
        _count += 1;
        emit CountChanged(msg.sender, _count);
    }

    /// @notice 计数增加指定值
    function incBy(uint256 delta) external {
        _count += delta;
        emit CountChanged(msg.sender, _count);
    }

    /// @notice 读取当前计数
    function get() external view returns (uint256) {
        return _count;
    }

    /// @notice 重置为 0，仅部署者可调用
    function reset() external {
        require(msg.sender == owner, "Counter: caller is not the owner");
        _count = 0;
        emit CountChanged(msg.sender, 0);
    }
}
