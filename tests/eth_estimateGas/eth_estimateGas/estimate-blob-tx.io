// checks gas estimation for blob transactions (EIP-4844)
// 
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"eth_estimateGas","params":[{"blobVersionedHashes":["0x0100000000000000000000000000000000000000000000000000000000000000"],"from":"0x0c2c51a0990aee1d73c1228de158688341557508","maxFeePerBlobGas":"0x5","nonce":"0x0","to":"0x0100000000000000000000000000000000000000","type":"0x5","value":"0x1"}]}
<< {"jsonrpc":"2.0","id":1,"result":"0x5208"}
