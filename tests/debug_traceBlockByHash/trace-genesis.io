// requests a trace of the genesis block by hash; must return an error since there is no parent state to replay from
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceBlockByHash","params":["0x202f4eff9f2f9447e497280d37eca304ec075091bc3c694f887cfe067f9cd43f"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"genesis is not traceable"}}
