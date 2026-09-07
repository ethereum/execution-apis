// traces a log-emitting contract call with the callTracer default config; logs must be absent without withLog
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceTransaction","params":["0xe972555e177d80d7b1bc96aa502df9c02c2da7d049236ee51f569d3c1e8c05ec",{"tracer":"callTracer"}]}
<< {"jsonrpc":"2.0","id":1,"result":{"from":"0x7435ed30a8b4aeb0877cef0c6e8cffe834eb865f","gas":"0x186a0","gasUsed":"0xca9c","to":"0x7dcd17433742f4c0ca53122ab541d0ba67fc27df","input":"0x9f5f8bad6b519820656d6974","value":"0x2","type":"CALL"}}
