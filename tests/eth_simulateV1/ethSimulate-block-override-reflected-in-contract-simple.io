// Checks that block overrides are true in contract for block number and time
>> {"jsonrpc":"2.0","id":1,"method":"eth_simulateV1","params":[{"blockStateCalls":[{"blockOverrides":{"number":"0x3b","time":"0x226"}},{"blockOverrides":{"number":"0x40","time":"0x230"}},{"blockOverrides":{"number":"0x4a","time":"0x23a"}}]},"latest"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-38021,"message":"block timestamps must be in order: 550 \u003c= 588"}}
