// Creates an access list for a zero-value transfer from an unfunded sender with all gas fee
// fields omitted. The client selects implementation-defined fee defaults for simulation, and the
// request must not fail merely because the sender cannot afford those defaults.
>> {"jsonrpc":"2.0","id":1,"method":"eth_createAccessList","params":[{"from":"0xaa00000000000000000000000000000000000000","to":"0xbb00000000000000000000000000000000000000"},"latest"]}
<< {"jsonrpc":"2.0","id":1,"result":{"accessList":[],"gasUsed":"0x5208"}}
