// traces a legacy EOA-to-EOA value transfer with the callTracer; a single frame with empty input and no calls or logs
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceTransaction","params":["0x3fbac8b19b59077cd29bbacc3815d73577b45a4d976cae80b04c98c793684c07",{"tracer":"callTracer"}]}
<< {"jsonrpc":"2.0","id":1,"result":{"from":"0x7435ed30a8b4aeb0877cef0c6e8cffe834eb865f","gas":"0x5208","gasUsed":"0x5208","to":"0xc7b99a164efd027a93f147376cc7da7c67c6bbe0","input":"0x","value":"0x1","type":"CALL"}}
