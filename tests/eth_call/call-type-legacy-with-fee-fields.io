// creates a contract that returns GASPRICE, with type 0x0 and the EIP-1559 fee fields.
// The client ignores type, so the fee fields set the gas price to the base fee plus the priority fee.
>> {"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","gas":"0x186a0","input":"0x3a60005260206000f3","maxFeePerGas":"0x77359400","maxPriorityFeePerGas":"0x3b9aca00","type":"0x0"},"latest"]}
<< {"jsonrpc":"2.0","id":1,"result":"0x000000000000000000000000000000000000000000000000000000003d3cdd97"}
