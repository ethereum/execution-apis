// traces a legacy EOA-to-EOA value transfer; structLogs must be empty since no EVM code runs
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceTransaction","params":["0x0d6999c0e9e4bec347945593e97bdcdf7c25be08ca1a1efdc520dbe75be985f3"]}
<< {"jsonrpc":"2.0","id":1,"result":{"gas":21000,"failed":false,"returnValue":"0x","structLogs":[]}}
