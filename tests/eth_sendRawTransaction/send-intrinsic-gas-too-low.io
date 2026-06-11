// sends a transaction with gas below the intrinsic minimum
>> {"jsonrpc":"2.0","id":1,"method":"eth_sendRawTransaction","params":["0xf868048405763d650194aa000000000000000000000000000000000000000a808718e5bb3abd109fa0c413e3d9e2011595cc27847e1a3e3549904d47384d7a71c9b9251fb845603e7fa0644eaaa2def4ca4a1c45c093090a28ea1067a9a4ca118c9dabbe4a8dbd1aecf1"]}
<< {"jsonrpc":"2.0","id":1,"error":{"code":800,"message":"intrinsic gas too low: gas 1, minimum needed 21000"}}
