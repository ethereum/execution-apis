// traces a legacy EOA-to-EOA value transfer; structLogs must be empty since no EVM code runs
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceTransaction","params":["0x3fbac8b19b59077cd29bbacc3815d73577b45a4d976cae80b04c98c793684c07"]}
<< {"jsonrpc":"2.0","id":1,"result":{"gas":21000,"failed":false,"returnValue":"0x","structLogs":[]}}
--

let method;
for (const m of messages) {
    if (m.send) {
        method = m.send.method;
        continue;
    }
    if (!method.startsWith("debug_trace")) continue;
    if (m.response.error) throw new Error("unexpected error response: " + JSON.stringify(m.response.error));
    const result = m.response.result;
    openrpc.validate("#/components/schemas/OpcodeTransactionTrace", result);
}
