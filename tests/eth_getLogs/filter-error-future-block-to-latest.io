// checks that an error is returned if `fromBlock` is greater than the latest block and `toBlock` is `latest`
>> {"jsonrpc":"2.0","id":1,"method":"eth_getLogs","params":[{"fromBlock":"0x38","toBlock":"latest"}]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"invalid block range params"}}
