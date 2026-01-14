// requests an invalid storage key
>> {"jsonrpc":"2.0","id":1,"method":"eth_getStorageAt","params":["0xaa00000000000000000000000000000000000000","0x00000000000000000000000000000000000000000000000000000000000000000","latest"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"unable to decode storage key: hex string too long, want at most 32 bytes"}}
