// estimates a transfer with type 0x3 and no blob fields on a block before Cancun. The client ignores type.
>> {"jsonrpc":"2.0","id":1,"method":"eth_estimateGas","params":[{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","to":"0x0100000000000000000000000000000000000000","type":"0x3"},"0x29"]}
<< {"jsonrpc":"2.0","id":1,"result":"0x5208"}
