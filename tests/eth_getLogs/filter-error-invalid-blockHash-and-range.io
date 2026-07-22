// checks that an error is returned if `fromBlock`/`toBlock` are specified together with `blockHash`
>> {"jsonrpc":"2.0","id":1,"method":"eth_getLogs","params":[{"blockHash":"0xf69b05b90b7e50c0b5b9b74d2d63a983dee56dffbbd68a530f026f263d76810c","fromBlock":"0x3","toBlock":"0x4"}]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"invalid argument 0: cannot specify both BlockHash and FromBlock/ToBlock, choose one or the other"}}
