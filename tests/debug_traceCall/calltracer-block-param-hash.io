// traces a call with the block given by hash
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceCall","params":[{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","gas":"0x100000","input":"0xff01","to":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d1"},"0xfbfbdd694e7ab88387234ef9ba3df04aef99b089870f5fd37fe4cb5998e5d4ce",{"tracer":"callTracer"}]}
<< {"jsonrpc":"2.0","id":1,"result":{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","gas":"0x100000","gasUsed":"0x5270","to":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d1","input":"0xff01","output":"0xffee","value":"0x0","type":"CALL"}}
