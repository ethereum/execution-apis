// checks that 32 zero-byte calldata meets the EIP-7623 floor
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"eth_estimateGas","params":[{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","input":"0x0000000000000000000000000000000000000000000000000000000000000000","nonce":"0x0","to":"0x0100000000000000000000000000000000000000","value":"0x1"}]}
<< {"jsonrpc":"2.0","id":1,"result":"0x53ee"}
