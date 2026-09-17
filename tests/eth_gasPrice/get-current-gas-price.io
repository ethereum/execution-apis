// gets the current gas price in wei
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"eth_gasPrice"}
<< {"jsonrpc":"2.0","id":1,"result":"0x1047435"}
--
if (BigInt(messages[1].response.result) <= 0) {
    throw new Error("gasprice too low");
}