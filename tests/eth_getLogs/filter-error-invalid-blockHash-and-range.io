// checks that an error is returned if `fromBlock`/`toBlock` are specified together with `blockHash`
>> {"jsonrpc":"2.0","id":1,"method":"eth_getLogs","params":[{"blockHash":"0xa54da6f03fa509ff00f062ff00ab1352614307374d7664c6e40d9764252dd3a0","fromBlock":"0x3","toBlock":"0x4"}]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"invalid argument 0: cannot specify both BlockHash and FromBlock/ToBlock, choose one or the other"}}
