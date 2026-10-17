// traces a call to the calltree contract with onlyTopCall; only the root frame is returned
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceCall","params":[{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","gas":"0x100000","to":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d0"},"latest",{"tracer":"callTracer","tracerConfig":{"onlyTopCall":true}}]}
<< {"jsonrpc":"2.0","id":1,"result":{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","gas":"0x100000","gasUsed":"0x4fd23","to":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d0","input":"0x","output":"0xffee","value":"0x0","type":"CALL"}}
