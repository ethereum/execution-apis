>> {"jsonrpc":"2.0","id":1,"method":"eth_simulateV1","params":[{"blockStateCalls":[{"stateOverrides":{"0xc000000000000000000000000000000000000000":{"balance":"0x4e20"}}},{"calls":[{"from":"0xc000000000000000000000000000000000000000","to":"0xc000000000000000000000000000000000000000","nonce":"0x0"},{"from":"0xc100000000000000000000000000000000000000","to":"0xc100000000000000000000000000000000000000","nonce":"0x1"},{"from":"0xc100000000000000000000000000000000000000","to":"0xc100000000000000000000000000000000000000","nonce":"0x0"}]}],"validation":true},"latest"]}
<< {"jsonrpc":"2.0","id":1,"result":[{"number":"0x4","hash":"0x4aca804390f3931f667f0ae572fb204881c5fba4ebd1ea735b6ccbfd38475d19","timestamp":"0x1f","gasLimit":"0x4c4b40","gasUsed":"0x0","feeRecipient":"0x0000000000000000000000000000000000000000","baseFeePerGas":"0x2310a91d","prevRandao":"0x0000000000000000000000000000000000000000000000000000000000000000","calls":[]},{"number":"0x5","hash":"0x6b1d7313f0b0078c737064915bf75ac42e796cfedd6a9d88462df43529eeb098","timestamp":"0x20","gasLimit":"0x4c4b40","gasUsed":"0xa410","feeRecipient":"0x0000000000000000000000000000000000000000","baseFeePerGas":"0x1eae93fa","prevRandao":"0x0000000000000000000000000000000000000000000000000000000000000000","calls":[{"returnData":"0x","logs":[],"gasUsed":"0x5208","status":"0x1"},{"returnData":"0x","logs":[],"gasUsed":"0x0","status":"0x0","error":{"message":"err: nonce too high: address 0xc100000000000000000000000000000000000000, tx: 1 state: 0 (supplied gas 4979000)","code":-38011}},{"returnData":"0x","logs":[],"gasUsed":"0x5208","status":"0x1"}]}]}
