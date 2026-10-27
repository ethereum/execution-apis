// calls with only maxPriorityFeePerGas on a block before London. The request is invalid because EIP-1559 is not active.
>> {"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","maxPriorityFeePerGas":"0x3b9aca00","to":"0x0100000000000000000000000000000000000000"},"0x1a"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"maxFeePerGas and maxPriorityFeePerGas are not valid before London is active"}}
