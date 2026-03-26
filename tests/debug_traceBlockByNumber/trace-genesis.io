// requests a trace of the genesis block; must return an error since there is no parent state to replay from
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceBlockByNumber","params":["0x0"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"genesis is not traceable"}}
