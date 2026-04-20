// retrieves pending transactions from a specific address
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"txpool_contentFrom","params":["0x0000000000000000000000000000000000000000"]}
<< {"jsonrpc":"2.0","id":1,"result":{"pending":{},"queued":{}}}
