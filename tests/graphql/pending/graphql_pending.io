>> mutation {sendRawTransaction(data: "0xf85021822710830186a080808200011ba01b733c6ee0c3ea21830991381618712eee6dd3a85136d3ac6c23b437ed6414e6a07709914716fb9751130f7c44f1d3ef94ceb02cbe0f6721608bb2050d96933e81")}
<< {"data":{"sendRawTransaction":"0x34b0f5643684a4379df8e81bd84207857d346eab44fd3f92cbbef9f72064fd35"}}

>> {pending {transactionCount transactions {nonce gas} account(address: "0x6295ee1b4f6dd65047762f924ecd367c17eabf8f") {balance} estimateGas(data: {}) call(data: {from: "0xa94f5374fce5edbc8e2a8697c15331677e6ebf0b" to: "0x6295ee1b4f6dd65047762f924ecd367c17eabf8f" data: "0x12a7b914"}) {data status}}}
<< {"data":{"pending":{"transactionCount":"0x1","transactions":[{"nonce":"0x21","gas":"0x186a0"}],"account":{"balance":"0x140"},"estimateGas":"0xcf08","call":{"data":"0x0000000000000000000000000000000000000000000000000000000000000001","status":"0x1"}}}}
