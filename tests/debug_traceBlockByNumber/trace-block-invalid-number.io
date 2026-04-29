// requests a trace with a non-hex block number; the client must return an error
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceBlockByNumber","params":["3"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"invalid argument 0: hex string without 0x prefix"}}
