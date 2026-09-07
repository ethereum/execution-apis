// traces a call at a non-existent block; an error is expected
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceCall","params":[{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","gas":"0x100000","to":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d1"},"0xfffffffff",{"tracer":"callTracer"}]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"block #68719476735 not found"}}
