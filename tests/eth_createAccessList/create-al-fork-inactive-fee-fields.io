// creates an access list with maxFeePerGas and maxPriorityFeePerGas on a block before London. The request is invalid because EIP-1559 is not active.
>> {"jsonrpc":"2.0","id":1,"method":"eth_createAccessList","params":[{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","maxFeePerGas":"0x77359400","maxPriorityFeePerGas":"0x3b9aca00","to":"0x0100000000000000000000000000000000000000"},"0x1a"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"maxFeePerGas and maxPriorityFeePerGas are not valid before London is active"}}
