// traces a contract creation whose initcode returns code that starts with 0xef with the callTracer, at the last block before London
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceCall","params":[{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","gas":"0x100000","input":"0x60ef60005360016000f3"},"0x1a",{"tracer":"callTracer"}]}
<< {"jsonrpc":"2.0","id":1,"result":{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","gas":"0x100000","gasUsed":"0xd06a","to":"0x8a700907005c5c3b442322abfb6bec88772ebd11","input":"0x60ef60005360016000f3","output":"0xef","value":"0x0","type":"CREATE"}}
