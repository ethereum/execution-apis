// traces a call whose child halts on a LOG0 in a static call; the failing step is the child's last entry, with an error and a gasCost of 0
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceCall","params":[{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","gas":"0x100000","to":"0x00000000000000000000000000000000000000c1"},"0x38",{"stateOverrides":{"0x00000000000000000000000000000000000000c1":{"code":"0x600060006000600060c262010000fa5000"},"0x00000000000000000000000000000000000000c2":{"code":"0x60006000a0"}}}]}
<< {"jsonrpc":"2.0","id":1,"result":{"gas":89156,"failed":false,"returnValue":"0x","structLogs":[{"pc":0,"op":"PUSH1","gas":1027576,"gasCost":3,"depth":1,"stack":[]},{"pc":2,"op":"PUSH1","gas":1027573,"gasCost":3,"depth":1,"stack":["0x0"]},{"pc":4,"op":"PUSH1","gas":1027570,"gasCost":3,"depth":1,"stack":["0x0","0x0"]},{"pc":6,"op":"PUSH1","gas":1027567,"gasCost":3,"depth":1,"stack":["0x0","0x0","0x0"]},{"pc":8,"op":"PUSH1","gas":1027564,"gasCost":3,"depth":1,"stack":["0x0","0x0","0x0","0x0"]},{"pc":10,"op":"PUSH3","gas":1027561,"gasCost":3,"depth":1,"stack":["0x0","0x0","0x0","0x0","0xc2"]},{"pc":14,"op":"STATICCALL","gas":1027558,"gasCost":68136,"depth":1,"stack":["0x0","0x0","0x0","0x0","0xc2","0x10000"]},{"pc":0,"op":"PUSH1","gas":65536,"gasCost":3,"depth":2,"stack":[]},{"pc":2,"op":"PUSH1","gas":65533,"gasCost":3,"depth":2,"stack":["0x0"]},{"pc":4,"op":"LOG0","gas":65530,"gasCost":0,"depth":2,"error":"write protection","stack":["0x0","0x0"]},{"pc":15,"op":"POP","gas":959422,"gasCost":2,"depth":1,"stack":["0x0"]},{"pc":16,"op":"STOP","gas":959420,"gasCost":0,"depth":1,"stack":[]}]}}
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
