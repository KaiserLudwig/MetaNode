// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title Voting 候选人投票合约
/// @notice 作业 1：用 mapping 记录候选人票数，支持投票、查询与重置
contract Voting {
    /// @notice 候选人名称 => 得票数
    mapping(string => uint256) private _votes;

    /// @notice 记录出现过的候选人，用于遍历与重置
    string[] private _candidates;

    /// @notice 候选人是否已登记
    mapping(string => bool) private _isCandidate;

    /// @param voter 投票人地址
    /// @param candidate 候选人名称
    /// @param totalVotes 该候选人累计票数
    event Voted(address indexed voter, string candidate, uint256 totalVotes);

    /// @param candidateCount 被重置的候选人数量
    event VotesReset(uint256 candidateCount);

    /// @notice 给某个候选人投一票；候选人首次出现时自动登记
    /// @dev 同一个地址可以投多次票，每调用一次计一票
    function vote(string memory candidate) public {
        require(bytes(candidate).length > 0, "Voting: empty candidate");

        if (!_isCandidate[candidate]) {
            _isCandidate[candidate] = true;
            _candidates.push(candidate);
        }

        _votes[candidate] += 1;
        emit Voted(msg.sender, candidate, _votes[candidate]);
    }

    /// @notice 查询某个候选人的得票数，未出现过的候选人返回 0
    function getVotes(string memory candidate) public view returns (uint256) {
        return _votes[candidate];
    }

    /// @notice 重置所有候选人的得票数（候选人列表保留）
    function resetVotes() public {
        uint256 count = _candidates.length;
        for (uint256 i = 0; i < count; i++) {
            _votes[_candidates[i]] = 0;
        }
        emit VotesReset(count);
    }

    /// @notice 返回全部候选人名称
    function candidates() public view returns (string[] memory) {
        return _candidates;
    }

    /// @notice 返回候选人数量
    function candidateCount() public view returns (uint256) {
        return _candidates.length;
    }
}
