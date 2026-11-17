// creates an access list for a transfer on a block before Berlin. The request is invalid because EIP-2930 is not active.
>> {"jsonrpc":"2.0","id":1,"method":"eth_createAccessList","params":[{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","to":"0x0100000000000000000000000000000000000000"},"0x17"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"failed to apply transaction: 0x6b16f58420ae01af4361f81985be0fe9b8c2400f9c247e093731d04609ff17e9 err: transaction type not supported: access list tx (sender 0x0c2c51a0990AeE1d73C1228de158688341557508)"}}
