// Creates an access list for a contract invocation that accesses storage.
// This invocation uses EIP-1559 fields to specify the gas price.
>> {"jsonrpc":"2.0","id":1,"method":"eth_createAccessList","params":[{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","gas":"0xea60","input":"0x010203040506","maxFeePerGas":"0x4426f2b","maxPriorityFeePerGas":"0x3","nonce":"0x0","to":"0x7dcd17433742f4c0ca53122ab541d0ba67fc27df"},"latest"]}
<< {"jsonrpc":"2.0","id":1,"result":{"accessList":[{"address":"0x7dcd17433742f4c0ca53122ab541d0ba67fc27df","storageKeys":["0x13a08e3cd39a1bc7bf9103f63f83273cced2beada9f723945176d6b983c65bd2","0x0000000000000000000000000000000000000000000000000000000000000000"]}],"gasUsed":"0xca3c"}}
