// Creates an access list for a contract invocation that reverts.
// The server should return the accessed slots regardless of failure, and should report the failure
// in the "error" field.
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"eth_createAccessList","params":[{"gas":"0x186a0","input":"0x01","to":"0x0ee3ab1371c93e7c0c281cc0c2107cdebc8b1930"},"latest"]}
<< {"jsonrpc":"2.0","id":1,"result":{"accessList":[{"address":"0x0ee3ab1371c93e7c0c281cc0c2107cdebc8b1930","storageKeys":["0x00000000000000000000000000000000000000000000000000000000000042ff"]}],"error":"execution reverted","gasUsed":"0x639d"}}
