// performs a contract call without the optional block parameter, which defaults to latest
>> {"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"input":"0xff01","to":"0x17e7eedce4ac02ef114a7ed9fe6e2f33feba1667"}]}
<< {"jsonrpc":"2.0","id":1,"result":"0xffee"}
