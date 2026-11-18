// traces an Amsterdam transaction whose SSTORE runs out of gas on the EIP-8037 state-gas charge; the failing step has one entry with an error and a gasCost of 0
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceTransaction","params":["0xfca19aa43410d6b3a7363c5dbc9e2c1a07c5c65979a74d48ddcb80b685520b3f"]}
<< {"jsonrpc":"2.0","id":1,"result":{"gas":100000,"failed":true,"returnValue":"0x","structLogs":[{"pc":0,"op":"CALLDATASIZE","gas":66532,"gasCost":2,"depth":1,"stack":[]},{"pc":1,"op":"DUP1","gas":66530,"gasCost":3,"depth":1,"stack":["0xc"]},{"pc":2,"op":"PUSH1","gas":66527,"gasCost":3,"depth":1,"stack":["0xc","0xc"]},{"pc":4,"op":"DUP1","gas":66524,"gasCost":3,"depth":1,"stack":["0xc","0xc","0x0"]},{"pc":5,"op":"CALLDATACOPY","gas":66521,"gasCost":9,"depth":1,"stack":["0xc","0xc","0x0","0x0"]},{"pc":6,"op":"PUSH1","gas":66512,"gasCost":3,"depth":1,"stack":["0xc"]},{"pc":8,"op":"KECCAK256","gas":66509,"gasCost":36,"depth":1,"stack":["0xc","0x0"]},{"pc":9,"op":"PUSH1","gas":66473,"gasCost":3,"depth":1,"stack":["0x93e49bbd0e156d6293b71176d5a32aedfc2b133b57811bd2762906ec17fff1a1"]},{"pc":11,"op":"SLOAD","gas":66470,"gasCost":100,"depth":1,"stack":["0x93e49bbd0e156d6293b71176d5a32aedfc2b133b57811bd2762906ec17fff1a1","0x0"],"storage":{"0x0000000000000000000000000000000000000000000000000000000000000000":"0x0000000000000000000000000000000000000000000000000000000000000033"}},{"pc":12,"op":"DUP1","gas":66370,"gasCost":3,"depth":1,"stack":["0x93e49bbd0e156d6293b71176d5a32aedfc2b133b57811bd2762906ec17fff1a1","0x33"]},{"pc":13,"op":"DUP3","gas":66367,"gasCost":3,"depth":1,"stack":["0x93e49bbd0e156d6293b71176d5a32aedfc2b133b57811bd2762906ec17fff1a1","0x33","0x33"]},{"pc":14,"op":"SSTORE","gas":66364,"gasCost":0,"depth":1,"error":"out of gas","stack":["0x93e49bbd0e156d6293b71176d5a32aedfc2b133b57811bd2762906ec17fff1a1","0x33","0x33","0x93e49bbd0e156d6293b71176d5a32aedfc2b133b57811bd2762906ec17fff1a1"]}]}}
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
