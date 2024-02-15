// gets fee history information
>> {"jsonrpc":"2.0","id":1,"method":"eth_feeHistory","params":["0x1","0x2",[95,99]]}
<< {"jsonrpc":"2.0","id":1,"result":{"oldestBlock":"0x2","reward":[["0x1","0x1"]],"baseFeePerBlobGas":["0x1","0x2"],"baseFeePerGas":["0x2dbf1f99","0x281d620d"],"blobGasUsedRatio":[0.5],"gasUsedRatio":[0.007565458319646006]}}
