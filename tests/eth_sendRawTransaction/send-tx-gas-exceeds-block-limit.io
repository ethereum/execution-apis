// sends a transaction with gasLimit exceeding the block gasLimit
>> {"jsonrpc":"2.0","id":1,"method":"eth_sendRawTransaction","params":["0xf86c048405763d6584047e7c4194aa000000000000000000000000000000000000000a808718e5bb3abd109fa052f7d99eba02d8b7e308b8d0882eae9af8a154770d790052651be01450f9533ba03c08c1f65cecd4a7e1cd218147e9fbcec1431a6e3f299723d093239ef819176d"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":803,"message":"exceeds block gas limit"}}
