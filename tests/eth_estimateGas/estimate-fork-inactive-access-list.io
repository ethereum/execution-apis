// estimates a transfer with an empty access list on a block before Berlin. The request is invalid because EIP-2930 is not active.
>> {"jsonrpc":"2.0","id":1,"method":"eth_estimateGas","params":[{"accessList":[],"from":"0x0c2c51a0990aee1d73c1228de158688341557508","to":"0x0100000000000000000000000000000000000000"},"0x17"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"failed with 50000000 gas: transaction type not supported: access list tx (sender 0x0c2c51a0990AeE1d73C1228de158688341557508)"}}
