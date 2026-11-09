// estimates a contract creation that ignores a failed nested creation; the estimate must leave the nested creation enough gas to succeed
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"eth_estimateGas","params":[{"from":"0x0102030000000000000000000000000000000000","input":"0x6c6175305b600190038060035700600052600d60136000f05000","to":null}]}
<< {"jsonrpc":"2.0","id":1,"result":"0x1261f1"}
--

if (BigInt(messages[1].response.result) < 1196078n) {
    throw new Error("gas estimate starves the nested creation");
}
