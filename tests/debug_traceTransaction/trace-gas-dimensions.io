// traces an Amsterdam transfer with the opcode logger; the result MUST carry the EIP-8037 executionGasUsed, stateGasUsed and gasRefund as integers
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceTransaction","params":["0xc5ab63ec1074b646fd6626aa56f3c8a30fd5ca43e8da8b011bc1d29279176742"]}
<< {"jsonrpc":"2.0","id":1,"result":{"gas":21000,"failed":false,"returnValue":"0x","structLogs":[],"executionGasUsed":21000,"stateGasUsed":0,"gasRefund":0}}
