// creates a contract that reads the balance of an address and returns the remaining gas, with type 0x0
// and an access list that contains the address. The client ignores type, so the result equals the second call,
// which sends the same request without type.
>> {"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"accessList":[{"address":"0x0100000000000000000000000000000000000000","storageKeys":[]}],"from":"0x0c2c51a0990aee1d73c1228de158688341557508","gas":"0x30d40","input":"0x73010000000000000000000000000000000000000031505a60005260206000f3","type":"0x0"},"latest"]}
<< {"jsonrpc":"2.0","id":1,"result":"0x0000000000000000000000000000000000000000000000000000000000023367"}
>> {"jsonrpc":"2.0","id":2,"method":"eth_call","params":[{"accessList":[{"address":"0x0100000000000000000000000000000000000000","storageKeys":[]}],"from":"0x0c2c51a0990aee1d73c1228de158688341557508","gas":"0x30d40","input":"0x73010000000000000000000000000000000000000031505a60005260206000f3"},"latest"]}
<< {"jsonrpc":"2.0","id":2,"result":"0x0000000000000000000000000000000000000000000000000000000000023367"}
