// traces a call whose child halts on an SSTORE that does not change its slot, with no more than the 2300 gas of the EIP-2200 sentry; the failing step is the child's last entry, with an error and a gasCost of 0
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceCall","params":[{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","gas":"0x100000","to":"0x00000000000000000000000000000000000000c1"},"0x38",{"stateOverrides":{"0x00000000000000000000000000000000000000c1":{"code":"0x6000600060006000600060c26108fcf15000"},"0x00000000000000000000000000000000000000c2":{"code":"0x6000600055"}}}]}
<< {"jsonrpc":"2.0","id":1,"result":{"gas":25923,"failed":false,"returnValue":"0x","structLogs":[{"pc":0,"op":"PUSH1","gas":1027576,"gasCost":3,"depth":1,"stack":[]},{"pc":2,"op":"PUSH1","gas":1027573,"gasCost":3,"depth":1,"stack":["0x0"]},{"pc":4,"op":"PUSH1","gas":1027570,"gasCost":3,"depth":1,"stack":["0x0","0x0"]},{"pc":6,"op":"PUSH1","gas":1027567,"gasCost":3,"depth":1,"stack":["0x0","0x0","0x0"]},{"pc":8,"op":"PUSH1","gas":1027564,"gasCost":3,"depth":1,"stack":["0x0","0x0","0x0","0x0"]},{"pc":10,"op":"PUSH1","gas":1027561,"gasCost":3,"depth":1,"stack":["0x0","0x0","0x0","0x0","0x0"]},{"pc":12,"op":"PUSH2","gas":1027558,"gasCost":3,"depth":1,"stack":["0x0","0x0","0x0","0x0","0x0","0xc2"]},{"pc":15,"op":"CALL","gas":1027555,"gasCost":4900,"depth":1,"stack":["0x0","0x0","0x0","0x0","0x0","0xc2","0x8fc"]},{"pc":0,"op":"PUSH1","gas":2300,"gasCost":3,"depth":2,"stack":[]},{"pc":2,"op":"PUSH1","gas":2297,"gasCost":3,"depth":2,"stack":["0x0"]},{"pc":4,"op":"SSTORE","gas":2294,"gasCost":0,"depth":2,"error":"out of gas: not enough gas for reentrancy sentry","stack":["0x0","0x0"]},{"pc":16,"op":"POP","gas":1022655,"gasCost":2,"depth":1,"stack":["0x0"]},{"pc":17,"op":"STOP","gas":1022653,"gasCost":0,"depth":1,"stack":[]}]}}
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
