// sends a transaction with maxPriorityFeePerGas greater than maxFeePerGas
>> {"jsonrpc":"2.0","id":1,"method":"eth_sendRawTransaction","params":["0x02f86b870c72dd9d5e883e048203e8018261a894aa000000000000000000000000000000000000000a80c080a07598a76d407d863e6dfb6a67a736782d0303fe47633fe28c1d7239d557cbd7e2a014f108ef0efc0234d5e64e4de30de849527b71cd1213eaaed1f99f29a0a9d3b5"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":804,"message":"max priority fee per gas higher than max fee per gas"}}
