// creates an access list for a transfer on a block before Berlin. The request is invalid because EIP-2930 is not active.
>> {"jsonrpc":"2.0","id":1,"method":"eth_createAccessList","params":[{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","to":"0x0100000000000000000000000000000000000000"},"0x17"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"failed to apply transaction: 0x93eac2a3162cccff521f874514811d908ea25414cbf11b0fd9bfe4f17f908381 err: transaction type not supported: access list tx (sender 0x0c2c51a0990AeE1d73C1228de158688341557508)"}}
