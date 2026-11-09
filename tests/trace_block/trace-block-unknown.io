// traces a block number past the head, which is unknown; the error is -32001
>> {"jsonrpc":"2.0","id":1,"method":"trace_block","params":["0x424"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32001,"message":"block 0x424 not found"}}
