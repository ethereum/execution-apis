// Moves the identity precompile (0x04) to 0x1000 while 0x1000 has its own override entry.
// The request must be rejected with error code -32000.
>> {"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"to":"0x00000000000000000000000000000000000c0de1"},"0x38",{"0x0000000000000000000000000000000000000004":{"movePrecompileToAddress":"0x0000000000000000000000000000000000001000"},"0x0000000000000000000000000000000000001000":{"balance":"0x1"},"0x00000000000000000000000000000000000c0de1":{"code":"0x00"}}]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"account 0x0000000000000000000000000000000000001000 is already overridden"}}
