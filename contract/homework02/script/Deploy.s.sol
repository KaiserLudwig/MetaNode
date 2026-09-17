// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {BeggingContract} from "../src/BeggingContract.sol";

interface VmBroadcast {
    function startBroadcast() external;
    function stopBroadcast() external;
}

address constant VM_ADDRESS = 0x7109709ECfa91a80626fF3989D68f67F5b1DD12D;

/// @notice 部署脚本：forge script script/Deploy.s.sol:DeployBegging --rpc-url $SEPOLIA_RPC_URL --broadcast
contract DeployBegging {
    VmBroadcast private constant vm = VmBroadcast(VM_ADDRESS);

    function run() external returns (BeggingContract deployed) {
        vm.startBroadcast();
        deployed = new BeggingContract();
        vm.stopBroadcast();
    }
}
