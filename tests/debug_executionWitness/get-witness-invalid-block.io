// requests the witness of a non-existent block; an error is expected
>> {"jsonrpc":"2.0","id":1,"method":"debug_executionWitness","params":["0xfffffffff"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"block 0xfffffffff not found"}}
