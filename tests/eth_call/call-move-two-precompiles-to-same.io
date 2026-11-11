// Moves ecMul (0x07) and identity (0x04) to 0x1000, with the 0x07 entry first in the
// request, then calls 0x1000 with 96 zero bytes and returns the size of the output. The precompile of
// the account with the higher address (ecMul, 64 bytes of output) must be at 0x1000, whatever the
// order of the entries; identity would return 96 bytes.
>> {"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"to":"0x00000000000000000000000000000000000c0de1"},"0x38",{"0x00000000000000000000000000000000000c0de1":{"code":"0x5f5f60605f6110005afa503d5f5260205ff3"},"0x0000000000000000000000000000000000000007":{"movePrecompileToAddress":"0x0000000000000000000000000000000000001000"},"0x0000000000000000000000000000000000000004":{"movePrecompileToAddress":"0x0000000000000000000000000000000000001000"}}]}
<< {"jsonrpc":"2.0","id":1,"result":"0x0000000000000000000000000000000000000000000000000000000000000040"}
