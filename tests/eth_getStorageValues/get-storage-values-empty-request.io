// requests with an empty map
>> {"jsonrpc":"2.0","id":1,"method":"eth_getStorageValues","params":[{},"latest"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"empty request"}}
