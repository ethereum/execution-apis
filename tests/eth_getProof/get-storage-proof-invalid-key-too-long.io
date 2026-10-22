// Requests a proof with a storage key of 33 bytes. The request is invalid.
>> {"jsonrpc":"2.0","id":1,"method":"eth_getProof","params":["0x7dcd17433742f4c0ca53122ab541d0ba67fc27df",["0x000000000000000000000000000000000000000000000000000000000000000001"],"latest"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"storage key too long (want at most 32 bytes): \"0x000000000000000000000000000000000000000000000000000000000000000001\""}}
