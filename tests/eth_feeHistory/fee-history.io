// gets fee history information
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"eth_feeHistory","params":["0x1","0x1b",[95,99]]}
<< {"jsonrpc":"2.0","id":1,"result":{"oldestBlock":"0x1b","reward":[["0x1","0x1"]],"baseFeePerGas":["0x3b9aca00","0x342a385a"],"gasUsedRatio":[0.00072868],"baseFeePerBlobGas":["0x0","0x0"],"blobGasUsedRatio":[0]}}
--
let r = messages[1].response.result;
if (r.reward.length != 1)
    throw new Error("expected exactly one reward");
if (r.reward[0][0] !== "0x1")
    throw new Error("wrong reward value in response");
