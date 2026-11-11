// Moves an address that is not a precompile. The request must be rejected with error code -32000.
>> {"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"to":"0x00000000000000000000000000000000000c0de1"},"0x38",{"0x0000000000000000000000000000000000005555":{"movePrecompileToAddress":"0x0000000000000000000000000000000000001000"},"0x00000000000000000000000000000000000c0de1":{"code":"0x00"}}]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"account 0x0000000000000000000000000000000000005555 is not a precompile"}}
