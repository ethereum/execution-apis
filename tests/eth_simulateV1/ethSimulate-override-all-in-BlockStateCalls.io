// override all values in block and see that they are set in return value
>> {"jsonrpc":"2.0","id":1,"method":"eth_simulateV1","params":[{"blockStateCalls":[{"blockOverrides":{"number":"0x40","time":"0x226","gasLimit":"0x3ec","feeRecipient":"0xc200000000000000000000000000000000000000","prevRandao":"0xc300000000000000000000000000000000000000000000000000000000000000","baseFeePerGas":"0x3ef"}}]},"latest"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-38021,"message":"block timestamps must be in order: 550 \u003c= 648"}}
