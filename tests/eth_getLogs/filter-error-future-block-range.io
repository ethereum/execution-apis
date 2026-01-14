// checks that an error is returned if `toBlock` is greater than the latest block
>> {"jsonrpc":"2.0","id":1,"method":"eth_getLogs","params":[{"address":null,"fromBlock":"0x29","toBlock":"0x2f","topics":null}]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"invalid block range params"}}
