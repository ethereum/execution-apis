// gets fee history information
>> {"jsonrpc":"2.0","id":1,"method":"eth_feeHistory","params":["0x1","0x2",[95,99]]}
<< {"jsonrpc":"2.0","id":1,"result":{"oldestBlock":"0x2","reward":[["0x1","0x1"]],"baseFeePerGas":["0x2dbf1f99","0x2824151f"],"gasUsedRatio":[0.009853708990006765],"baseFeePerBlobGas":["0x1","0x1"],"blobGasUsedRatio":[0]}}
