// requests a trace of the genesis block by hash; must return an error since there is no parent state to replay from
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceBlockByHash","params":["0x79e0f5ffb6a0c9f54d507dbbadc935603ac1db86d32f7857472180a75ec11f90"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"genesis is not traceable"}}
