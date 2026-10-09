// estimates a transfer with maxFeePerBlobGas and no blob hashes on a block before Cancun. The request is invalid because EIP-4844 is not active.
>> {"jsonrpc":"2.0","id":1,"method":"eth_estimateGas","params":[{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","maxFeePerBlobGas":"0x1","to":"0x0100000000000000000000000000000000000000"},"0x29"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"failed with 50000000 gas: transaction type not supported: blob tx (sender 0x0c2c51a0990AeE1d73C1228de158688341557508)"}}
