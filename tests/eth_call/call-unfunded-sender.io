// Performs a call with a zero-value transfer from an unfunded sender with all gas fee
// fields omitted. The request must not fail merely because the sender cannot afford the client's fee defaults.
>> {"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"from":"0xaa00000000000000000000000000000000000000","to":"0xbb00000000000000000000000000000000000000"},"latest"]}
<< {"jsonrpc":"2.0","id":1,"result":"0x"}
