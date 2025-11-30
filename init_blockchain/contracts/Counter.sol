// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

// 计数器合约
contract Counter {
    // 状态变量：计数器值
    uint256 private count;

    // 事件：当计数器值改变时触发
    event CountChanged(uint256 oldValue, uint256 newValue);

    // 构造函数：初始化计数器
    constructor(uint256 _initialCount) {
        count = _initialCount;
    }

    // 获取当前计数器值
    function getCount() public view returns (uint256) {
        return count;
    }

    // 增加计数器
    function increment() public {
        uint256 oldValue = count;
        count += 1;
        emit CountChanged(oldValue, count);
    }

    // 减少计数器
    function decrement() public {
        require(count > 0, "Counter: cannot decrement below zero");
        uint256 oldValue = count;
        count -= 1;
        emit CountChanged(oldValue, count);
    }

    // 设置计数器值
    function setCount(uint256 _newCount) public {
        uint256 oldValue = count;
        count = _newCount;
        emit CountChanged(oldValue, count);
    }
}