// Accepts targetGasLimit 0x0 and starts payload building at the known Amsterdam head. The validation script checks the client-specific payloadId as 8-byte DATA and compares all other fields exactly.
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"engine_forkchoiceUpdatedV4","params":[{"finalizedBlockHash":"0x52bccfc9c4d4874b1898c0fbd0ae3b598a0edc7b94d9b796857dbbe1093c32e0","headBlockHash":"0x52bccfc9c4d4874b1898c0fbd0ae3b598a0edc7b94d9b796857dbbe1093c32e0","safeBlockHash":"0x52bccfc9c4d4874b1898c0fbd0ae3b598a0edc7b94d9b796857dbbe1093c32e0"},{"parentBeaconBlockRoot":"0x0000000000000000000000000000000000000000000000000000000000000000","prevRandao":"0x0000000000000000000000000000000000000000000000000000000000000000","slotNumber":"0x4","suggestedFeeRecipient":"0x0000000000000000000000000000000000000000","targetGasLimit":"0x0","timestamp":"0x264","withdrawals":[]},null]}
<< {"jsonrpc":"2.0","id":1,"result":{"payloadStatus":{"status":"VALID","latestValidHash":"0x52bccfc9c4d4874b1898c0fbd0ae3b598a0edc7b94d9b796857dbbe1093c32e0","validationError":null},"payloadId":"0x043d1b3020ebb6c5"}}
--

const expected = messages[1].expected;
const response = messages[1].response;
const id = response.result && response.result.payloadId;
if (typeof id !== "string" || !/^0x[0-9a-f]{16}$/.test(id)) {
    throw new Error("expected a non-null 8-byte payloadId");
}
response.result.payloadId = expected.result.payloadId;
assertEqual(response, expected, "response");

function assertEqual(actual, expected, path) {
    if (actual === expected) return;
    if (actual === null || expected === null ||
        typeof actual !== "object" || typeof expected !== "object" ||
        Array.isArray(actual) !== Array.isArray(expected)) {
        throw new Error(path + " differs from the expected response");
    }
    const actualKeys = Object.keys(actual).sort();
    const expectedKeys = Object.keys(expected).sort();
    if (JSON.stringify(actualKeys) !== JSON.stringify(expectedKeys)) {
        throw new Error(path + " has unexpected or missing fields");
    }
    for (const key of expectedKeys) {
        assertEqual(actual[key], expected[key], path + "." + key);
    }
}
