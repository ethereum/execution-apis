// traces a calltree contract invocation with onlyTopCall; only the root frame is returned
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceTransaction","params":["0xa3cec99b572cbb16af9a680a23b7925523b286df43c56548f52f3b1677c99189",{"tracer":"callTracer","tracerConfig":{"onlyTopCall":true}}]}
<< {"jsonrpc":"2.0","id":1,"result":{"from":"0x7435ed30a8b4aeb0877cef0c6e8cffe834eb865f","gas":"0x927c0","gasUsed":"0x28645","to":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d0","input":"0x","output":"0xffee","value":"0x0","type":"CALL"}}
