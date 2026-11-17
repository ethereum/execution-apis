// creates an access list with an empty blobVersionedHashes list and no maxFeePerBlobGas. The request is invalid at every fork.
>> {"jsonrpc":"2.0","id":1,"method":"eth_createAccessList","params":[{"blobVersionedHashes":[],"from":"0x0c2c51a0990aee1d73c1228de158688341557508","to":"0x0100000000000000000000000000000000000000"},"latest"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"failed to apply transaction: 0x9a00d084acfd261c20f21a94f01e7de83b2e584d41091158716e8cd3e3cac9c2 err: blob transaction missing blob hashes"}}
