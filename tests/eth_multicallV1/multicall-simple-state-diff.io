>> {"jsonrpc":"2.0","id":1,"method":"eth_simulateV1","params":[{"blockStateCalls":[{"stateOverrides":{"0xc000000000000000000000000000000000000000":{"balance":"0x7d0"},"0xc100000000000000000000000000000000000000":{"code":"0x608060405234801561001057600080fd5b506004361061004c5760003560e01c80630ff4c916146100515780633033413b1461008157806344e12f871461009f5780637b8d56e3146100bd575b600080fd5b61006b600480360381019061006691906101f6565b6100d9565b6040516100789190610232565b60405180910390f35b61008961013f565b6040516100969190610232565b60405180910390f35b6100a7610145565b6040516100b49190610232565b60405180910390f35b6100d760048036038101906100d2919061024d565b61014b565b005b60006002821061011e576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610115906102ea565b60405180910390fd5b6000820361012c5760005490505b6001820361013a5760015490505b919050565b60015481565b60005481565b6002821061018e576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610185906102ea565b60405180910390fd5b600082036101a257806000819055506101b7565b600182036101b657806001819055506101b7565b5b5050565b600080fd5b6000819050919050565b6101d3816101c0565b81146101de57600080fd5b50565b6000813590506101f0816101ca565b92915050565b60006020828403121561020c5761020b6101bb565b5b600061021a848285016101e1565b91505092915050565b61022c816101c0565b82525050565b60006020820190506102476000830184610223565b92915050565b60008060408385031215610264576102636101bb565b5b6000610272858286016101e1565b9250506020610283858286016101e1565b9150509250929050565b600082825260208201905092915050565b7f746f6f2062696720736c6f740000000000000000000000000000000000000000600082015250565b60006102d4600c8361028d565b91506102df8261029e565b602082019050919050565b60006020820190508181036000830152610303816102c7565b905091905056fea2646970667358221220ceea194bb66b5b9f52c83e5bf5a1989255de8cb7157838eff98f970c3a04cb3064736f6c63430008120033"}},"calls":[{"from":"0xc000000000000000000000000000000000000000","to":"0xc100000000000000000000000000000000000000","input":"0x7b8d56e300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"},{"from":"0xc000000000000000000000000000000000000000","to":"0xc100000000000000000000000000000000000000","input":"0x7b8d56e300000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002"}]},{"stateOverrides":{"0xc100000000000000000000000000000000000000":{"state":{"0x0000000000000000000000000000000000000000000000000000000000000000":"0x1200000000000000000000000000000000000000000000000000000000000000"}}},"calls":[{"from":"0xc000000000000000000000000000000000000000","to":"0xc100000000000000000000000000000000000000","input":"0x0ff4c9160000000000000000000000000000000000000000000000000000000000000000"},{"from":"0xc000000000000000000000000000000000000000","to":"0xc100000000000000000000000000000000000000","input":"0x0ff4c9160000000000000000000000000000000000000000000000000000000000000001"}]}],"traceTransfers":true},"latest"]}
<< {"jsonrpc":"2.0","id":1,"result":[{"number":"0x4","hash":"0x6093bd45b013eaf05790f67641d02029efae7ff806f3f6681e81a0f4b26b2e3c","timestamp":"0x1f","gasLimit":"0x4c4b40","gasUsed":"0x158df","feeRecipient":"0x0000000000000000000000000000000000000000","baseFeePerGas":"0x2310a91d","prevRandao":"0x0000000000000000000000000000000000000000000000000000000000000000","calls":[{"returnData":"0x","logs":[],"gasUsed":"0xac5e","status":"0x1"},{"returnData":"0x","logs":[],"gasUsed":"0xac81","status":"0x1"}]},{"number":"0x5","hash":"0xcab90b9b7f7a82440d66d354174348d5130450bb873f8d2263a3e647a224049f","timestamp":"0x20","gasLimit":"0x4c4b40","gasUsed":"0xbb16","feeRecipient":"0x0000000000000000000000000000000000000000","baseFeePerGas":"0x1eae93fa","prevRandao":"0x0000000000000000000000000000000000000000000000000000000000000000","calls":[{"returnData":"0x1200000000000000000000000000000000000000000000000000000000000000","logs":[],"gasUsed":"0x5d85","status":"0x1"},{"returnData":"0x0000000000000000000000000000000000000000000000000000000000000000","logs":[],"gasUsed":"0x5d91","status":"0x1"}]}]}