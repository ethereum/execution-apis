// traces a call whose child allocates a storage slot with a gas limit below the EIP-7825 cap; the child's gasUsed includes the state gas drawn from gas_left
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceCall","params":[{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","gas":"0x100000","to":"0x00000000000000000000000000000000000000c1"},"latest",{"stateOverrides":{"0x00000000000000000000000000000000000000c1":{"code":"0x5f5f5f5f5f60c25af15000","state":{}},"0x00000000000000000000000000000000000000c2":{"code":"0x600160015500","state":{}}},"tracer":"callTracer"}]}
<< {"jsonrpc":"2.0","id":1,"result":{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","gas":"0x100000","gasUsed":"0x1f42b","to":"0x00000000000000000000000000000000000000c1","input":"0x","calls":[{"from":"0x00000000000000000000000000000000000000c1","gas":"0xf7abb","gasUsed":"0x1adca","to":"0x00000000000000000000000000000000000000c2","input":"0x","value":"0x0","type":"CALL"}],"value":"0x0","type":"CALL"}}
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
    openrpc.validate("#/components/schemas/CallFrame", result);
    const d = diff(mask(m.expected.result), mask(result), "result");
    if (d) throw new Error("response differs from the fixture at " + d);
}

// Clears what the spec leaves to the client: callTracer error text other than
// REVERT and the text of a Panic revertReason.
function mask(v) {
    if (Array.isArray(v)) {
        v.forEach(mask);
    } else if (v !== null && typeof v === "object") {
        if (typeof v.error === "string" && v.error !== "execution reverted") v.error = "*";
        if (typeof v.revertReason === "string" && /^0x4e487b71/.test(v.output)) v.revertReason = "*";
        Object.values(v).forEach(mask);
    }
    return v;
}

function diff(want, got, path) {
    if (Array.isArray(want) || Array.isArray(got)) {
        if (!Array.isArray(want) || !Array.isArray(got)) return path + ": got " + JSON.stringify(got);
        if (want.length !== got.length) return path + ": got " + got.length + " items, want " + want.length;
        for (let i = 0; i < want.length; i++) {
            const d = diff(want[i], got[i], path + "[" + i + "]");
            if (d) return d;
        }
        return "";
    }
    if (want !== null && got !== null && typeof want === "object" && typeof got === "object") {
        for (const k of new Set(Object.keys(want).concat(Object.keys(got)))) {
            if (!(k in got)) return path + "." + k + ": missing";
            if (!(k in want)) return path + "." + k + ": unexpected field";
            const d = diff(want[k], got[k], path + "." + k);
            if (d) return d;
        }
        return "";
    }
    return want === got ? "" : path + ": got " + JSON.stringify(got) + ", want " + JSON.stringify(want);
}
