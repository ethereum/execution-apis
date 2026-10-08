// sends a transaction with a nonce below the sender's state nonce
>> {"jsonrpc":"2.0","id":1,"method":"eth_sendRawTransaction","params":["0xf86a808405763d658261a894aa000000000000000000000000000000000000000a808718e5bb3abd10a0a05f7f21951b14d685214b378a8d430f72e037ca02004712cbfc75ed124d06547da01f33468ef4837ef389a72fa6b1d6d89b97196f39e96f493ce7f3e3161a577202"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":1,"message":"nonce too low: next nonce 160, tx nonce 0"}}
