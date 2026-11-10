// Requests a proof with a storage key that is not hex. The request is invalid.
>> {"jsonrpc":"2.0","id":1,"method":"eth_getProof","params":["0x7dcd17433742f4c0ca53122ab541d0ba67fc27df",["0xzz"],"latest"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"invalid hex in storage key: \"0xzz\""}}
