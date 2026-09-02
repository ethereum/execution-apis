// requests a trace of the genesis block by hash; must return an error since there is no parent state to replay from
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceBlockByHash","params":["0x44fd89d504659cd58f48f4796b77a7e7012cf296a2409afa2f6c3cb99b5b3d99"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"genesis is not traceable"}}
