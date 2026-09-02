// checks that an error is returned if `fromBlock` is larger than `toBlock`
>> {"jsonrpc":"2.0","id":1,"method":"eth_getLogs","params":[{"fromBlock":"0x32","toBlock":"0x2f"}]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"invalid block range params"}}
