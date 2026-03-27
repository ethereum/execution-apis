// requests a trace for a non-existent block hash; the client must return an error
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceBlockByHash","params":["0x0000000000000000000000000000000000000000000000000000000000000001"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"block 0x0000000000000000000000000000000000000000000000000000000000000001 not found"}}
