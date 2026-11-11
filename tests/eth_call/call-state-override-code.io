// Calls an address with no code on chain, after a state override gives it code that
// returns 0x2a. The override must apply.
>> {"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"to":"0x00000000000000000000000000000000000c0de1"},"0x38",{"0x00000000000000000000000000000000000c0de1":{"code":"0x602a5f5260205ff3"}}]}
<< {"jsonrpc":"2.0","id":1,"result":"0x000000000000000000000000000000000000000000000000000000000000002a"}
