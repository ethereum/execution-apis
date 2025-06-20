// checks that including ephemeral authorizations increases gas
// 
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"eth_estimateGas","params":[{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","nonce":"0x0","to":"0x0100000000000000000000000000000000000000","value":"0x1"}]}
<< {"jsonrpc":"2.0","id":1,"result":"0x5208"}
>> {"jsonrpc":"2.0","id":2,"method":"eth_estimateGas","params":[{"authorizationList":[{"address":"0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","chainId":"0x1","nonce":"0x0","r":"0x1111111111111111111111111111111111111111111111111111111111111111","s":"0x2222222222222222222222222222222222222222222222222222222222222222","yParity":"0x0"}],"from":"0x0c2c51a0990aee1d73c1228de158688341557508","nonce":"0x0","to":"0x0100000000000000000000000000000000000000","type":"0x4","value":"0x1"}]}
<< {"jsonrpc":"2.0","id":2,"result":"0xb52e"}
