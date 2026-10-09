// calls with type 0x1 and no access list on a block before Berlin. The client ignores type, so the call runs as a legacy call.
>> {"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","to":"0x0100000000000000000000000000000000000000","type":"0x1"},"0x17"]}
<< {"jsonrpc":"2.0","id":1,"result":"0x"}
