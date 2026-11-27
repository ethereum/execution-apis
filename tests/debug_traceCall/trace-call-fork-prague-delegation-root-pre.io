// traces a call to an account with an EIP-7702 delegation designator with the default struct logger, at the last block before Prague
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceCall","params":[{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","gas":"0x100000","to":"0x00000000000000000000000000000000000000e7"},"0x2c",{"stateOverrides":{"0x00000000000000000000000000000000000000c3":{"code":"0x3060005260aa60206000a160206000f3"},"0x00000000000000000000000000000000000000e7":{"code":"0xef010000000000000000000000000000000000000000c3"}}}]}
<< {"jsonrpc":"2.0","id":1,"result":{"gas":1048576,"failed":true,"returnValue":"0x","structLogs":[{"pc":0,"op":"opcode 0xef not defined","gas":1027576,"gasCost":0,"depth":1,"error":"invalid opcode: opcode 0xef not defined","stack":[]}]}}
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
    openrpc.validate("#/components/schemas/OpcodeTransactionTrace", result);
    const d = diff(m.expected.result, result, "result");
    if (d) throw new Error("response differs from the fixture at " + d);
}

function errorText(v) {
    return typeof v === "string" && v !== "";
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
        const step = /\.structLogs\[\d+\]$/.test(path);
        for (const k of new Set(Object.keys(want).concat(Object.keys(got)))) {
            if (!(k in got)) return path + "." + k + ": missing";
            if (!(k in want)) return path + "." + k + ": unexpected field";
            if (step && k === "error" && errorText(want[k]) && errorText(got[k])) continue;
            const d = diff(want[k], got[k], path + "." + k);
            if (d) return d;
        }
        return "";
    }
    return want === got ? "" : path + ": got " + JSON.stringify(got) + ", want " + JSON.stringify(want);
}
