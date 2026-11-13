// checks that an error is returned if both `fromBlock` and `toBlock` are greater than the latest block
>> {"jsonrpc":"2.0","id":1,"method":"eth_getLogs","params":[{"fromBlock":"0x3e","toBlock":"0x40"}]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"block range extends beyond current head block"}}
