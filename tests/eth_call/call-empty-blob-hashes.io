// calls with an empty blobVersionedHashes list and no maxFeePerBlobGas. The request is invalid at every fork.
>> {"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"blobVersionedHashes":[],"from":"0x0c2c51a0990aee1d73c1228de158688341557508","to":"0x0100000000000000000000000000000000000000"},"latest"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"err: blob transaction missing blob hashes (supplied gas 50000000)"}}
