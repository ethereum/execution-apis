// traces a legacy EOA-to-EOA value transfer; structLogs must be empty since no EVM code runs
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceTransaction","params":["0x3fbac8b19b59077cd29bbacc3815d73577b45a4d976cae80b04c98c793684c07"]}
<< {"jsonrpc":"2.0","id":1,"result":{"gas":21000,"failed":false,"returnValue":"0x","structLogs":[]}}
