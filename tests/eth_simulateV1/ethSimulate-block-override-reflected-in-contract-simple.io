// Checks that block overrides are true in contract for block number and time
>> {"jsonrpc":"2.0","id":1,"method":"eth_simulateV1","params":[{"blockStateCalls":[{"blockOverrides":{"number":"0x41","time":"0x262"}},{"blockOverrides":{"number":"0x46","time":"0x26c"}},{"blockOverrides":{"number":"0x50","time":"0x276"}}]},"latest"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-38021,"message":"block timestamps must be in order: 610 \u003c= 648"}}
