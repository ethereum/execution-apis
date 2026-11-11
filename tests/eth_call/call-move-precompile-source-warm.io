// Moves the identity precompile (0x04) to 0x1000 and returns the gas that a STATICCALL
// to 0x04 costs. 0x04 stays warm and has no code, so the call pays only the warm access cost (100).
>> {"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"to":"0x00000000000000000000000000000000000c0de1"},"0x38",{"0x0000000000000000000000000000000000000004":{"movePrecompileToAddress":"0x0000000000000000000000000000000000001000"},"0x00000000000000000000000000000000000c0de1":{"code":"0x5a5f5f5f5f60045afa505a90035f5260205ff3"}}]}
<< {"jsonrpc":"2.0","id":1,"result":"0x0000000000000000000000000000000000000000000000000000000000000075"}
