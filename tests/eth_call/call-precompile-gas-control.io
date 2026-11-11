// Returns the gas that a STATICCALL to the identity precompile (0x04) costs, with no
// move. The call pays the warm access cost (100) and the identity cost (15).
>> {"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"to":"0x00000000000000000000000000000000000c0de1"},"0x38",{"0x00000000000000000000000000000000000c0de1":{"code":"0x5a5f5f5f5f60045afa505a90035f5260205ff3"}}]}
<< {"jsonrpc":"2.0","id":1,"result":"0x0000000000000000000000000000000000000000000000000000000000000084"}
