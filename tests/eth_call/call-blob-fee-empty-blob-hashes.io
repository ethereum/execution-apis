// calls with maxFeePerBlobGas and an empty blobVersionedHashes list. The request is invalid at every fork.
>> {"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"blobVersionedHashes":[],"from":"0x0c2c51a0990aee1d73c1228de158688341557508","maxFeePerBlobGas":"0x1","to":"0x0100000000000000000000000000000000000000"},"latest"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"maxFeePerBlobGas specified without blobVersionedHashes"}}
