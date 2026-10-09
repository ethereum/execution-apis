// creates an access list for a transfer from an unfunded sender with maxFeePerBlobGas and no blob hashes. From Cancun the field does not make the request invalid, and it does not need funds from the sender.
>> {"jsonrpc":"2.0","id":1,"method":"eth_createAccessList","params":[{"from":"0xaa00000000000000000000000000000000000000","maxFeePerBlobGas":"0x1","to":"0x0100000000000000000000000000000000000000"},"latest"]}
<< {"jsonrpc":"2.0","id":1,"result":{"accessList":[],"gasUsed":"0x5208"}}
