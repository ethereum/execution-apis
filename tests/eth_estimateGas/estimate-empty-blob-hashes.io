// estimates a transfer with an empty blobVersionedHashes list and no maxFeePerBlobGas. The request is invalid at every fork.
>> {"jsonrpc":"2.0","id":1,"method":"eth_estimateGas","params":[{"blobVersionedHashes":[],"from":"0x0c2c51a0990aee1d73c1228de158688341557508","to":"0x0100000000000000000000000000000000000000"},"latest"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"failed with 50000000 gas: blob transaction missing blob hashes"}}
