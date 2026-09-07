// traces a call with the block given by hash
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceCall","params":[{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","gas":"0x100000","input":"0xff01","to":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d1"},"0x52bccfc9c4d4874b1898c0fbd0ae3b598a0edc7b94d9b796857dbbe1093c32e0",{"tracer":"callTracer"}]}
<< {"jsonrpc":"2.0","id":1,"result":{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","gas":"0x100000","gasUsed":"0x3b18","to":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d1","input":"0xff01","output":"0xffee","value":"0x0","type":"CALL"}}
