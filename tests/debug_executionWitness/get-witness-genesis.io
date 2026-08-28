// requests the witness of the genesis block; an error is expected because genesis has no parent header
>> {"jsonrpc":"2.0","id":1,"method":"debug_executionWitness","params":["0x0"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"block 0x0 found, but parent missing"}}
