// traces a call with an empty withdrawals block override; the client rejects it with -32602
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceCall","params":[{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","gas":"0x100000","input":"0xff01","to":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d1"},"0xbdb683cb54c75b2851727aca9197d84a4c72bfaf6586b1afbd95c3bc82ab770e",{"blockOverrides":{"withdrawals":[]},"tracer":"callTracer"}]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"block override \"withdrawals\" is not supported for this RPC method"}}
