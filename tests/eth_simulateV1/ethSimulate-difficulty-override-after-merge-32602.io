// rejects a difficulty override for a block after the merge (-32602)
>> {"jsonrpc":"2.0","id":1,"method":"eth_simulateV1","params":[{"blockStateCalls":[{"blockOverrides":{"difficulty":"0x1"},"calls":[{"from":"0xc000000000000000000000000000000000000000","to":"0xc100000000000000000000000000000000000000"}]}]},"latest"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"block 61: difficulty override is not allowed after the merge"}}
