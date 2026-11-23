// rejects a block override key that the spec does not define (-32602)
>> {"jsonrpc":"2.0","id":1,"method":"eth_simulateV1","params":[{"blockStateCalls":[{"blockOverrides":{"bogusField":"0x1"},"calls":[{"from":"0xc000000000000000000000000000000000000000","to":"0xc100000000000000000000000000000000000000"}]}]},"latest"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"invalid argument 0: unknown field \"bogusField\" in block override"}}
