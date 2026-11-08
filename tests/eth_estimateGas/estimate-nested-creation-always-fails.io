// estimates a contract creation that ignores a nested creation which fails at any gas limit; the estimate must not grow because of it
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"eth_estimateGas","params":[{"from":"0x0102030000000000000000000000000000000000","input":"0x63600080fd6000526004601c6000f05000","to":null}]}
<< {"jsonrpc":"2.0","id":1,"result":"0x63013"}
--

if (BigInt(messages[1].response.result) >= 1000000n) {
    throw new Error("gas estimate grew on a nested call that always fails");
}
