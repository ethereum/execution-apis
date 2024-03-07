>> {"jsonrpc":"2.0","id":1,"method":"eth_simulateV1","params":[{"blockStateCalls":[{"stateOverrides":{"0xc000000000000000000000000000000000000000":{"code":"0x60806040526004361061001e5760003560e01c80634b64e49214610023575b600080fd5b61003d6004803603810190610038919061011f565b61003f565b005b60008173ffffffffffffffffffffffffffffffffffffffff166108fc349081150290604051600060405180830381858888f193505050509050806100b8576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016100af906101a9565b60405180910390fd5b5050565b600080fd5b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b60006100ec826100c1565b9050919050565b6100fc816100e1565b811461010757600080fd5b50565b600081359050610119816100f3565b92915050565b600060208284031215610135576101346100bc565b5b60006101438482850161010a565b91505092915050565b600082825260208201905092915050565b7f4661696c656420746f2073656e64204574686572000000000000000000000000600082015250565b600061019360148361014c565b915061019e8261015d565b602082019050919050565b600060208201905081810360008301526101c281610186565b905091905056fea2646970667358221220563acd6f5b8ad06a3faf5c27fddd0ecbc198408b99290ce50d15c2cf7043694964736f6c63430008120033","balance":"0x3e8"}},"calls":[{"from":"0xc000000000000000000000000000000000000000","to":"0xc100000000000000000000000000000000000000","value":"0x3e8"}]}],"traceTransfers":true,"validation":true},"latest"]}
<< {"jsonrpc":"2.0","id":1,"result":[{"number":"0x4","hash":"0x1827accb2edd8466cbacbe8353a6ab880dad99fd77bb9c0ca3616d2367faf2b9","timestamp":"0x1f","gasLimit":"0x4c4b40","gasUsed":"0x0","feeRecipient":"0x0000000000000000000000000000000000000000","baseFeePerGas":"0x2310a91d","prevRandao":"0x0000000000000000000000000000000000000000000000000000000000000000","calls":[{"returnData":"0x","logs":[],"gasUsed":"0x0","status":"0x0","error":{"message":"err: sender not an eoa: address 0xC000000000000000000000000000000000000000, codehash: 0x3b4b17870a84d1ab8ac96e2596ba4a999f2dcd3c73a622b5c308a031fe89d005 (supplied gas 5000000)","code":-32603}}]}]}