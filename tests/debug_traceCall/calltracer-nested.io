// traces a call to the calltree contract with the callTracer; the frame tree must contain every call type
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceCall","params":[{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","gas":"0x100000","to":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d0"},"latest",{"tracer":"callTracer"}]}
<< {"jsonrpc":"2.0","id":1,"result":{"from":"0x0c2c51a0990aee1d73c1228de158688341557508","gas":"0x100000","gasUsed":"0x4fd23","to":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d0","input":"0x","output":"0xffee","calls":[{"from":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d0","gas":"0xf35c","gasUsed":"0x48","to":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d1","input":"0xff01","output":"0xffee","value":"0x1","type":"CALL"},{"from":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d0","gas":"0xea60","gasUsed":"0x889","to":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d3","input":"0x0000000000000000000000000000000000000000000000000000000000000001","output":"0x08c379a00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000a75736572206572726f72","error":"execution reverted","revertReason":"user error","value":"0x0","type":"CALL"},{"from":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d0","gas":"0xea60","gasUsed":"0x7d","to":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d2","input":"0x","output":"0x000000000000000000000000000000000000000000000000000000000000003c000000000000000000000000000000000000000000000000000c72dd9d5e883e0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c2c51a0990aee1d73c1228de1586883415575080000000000000000000000000000000000000000000000000000000000000000","type":"STATICCALL"},{"from":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d0","gas":"0xea60","gasUsed":"0xea60","to":"0x7dcd17433742f4c0ca53122ab541d0ba67fc27df","input":"0x","error":"out of gas: write protection","type":"STATICCALL"},{"from":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d0","gas":"0xea60","gasUsed":"0x64b0","to":"0x7dcd17433742f4c0ca53122ab541d0ba67fc27df","input":"0x00000000","value":"0x0","type":"DELEGATECALL"},{"from":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d0","gas":"0xea60","gasUsed":"0x48","to":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d1","input":"0xff01","output":"0xffee","value":"0x0","type":"CALLCODE"},{"from":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d0","gas":"0xea60","gasUsed":"0x12","to":"0x0000000000000000000000000000000000000004","input":"0xff010000","output":"0xff010000","value":"0x0","type":"CALL"},{"from":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d0","gas":"0xaefed","gasUsed":"0x1682","to":"0xcea903b7562cb720503a8ee443d37367587ee3d0","input":"0x5b646368696c6460006000a133ff","calls":[{"from":"0xcea903b7562cb720503a8ee443d37367587ee3d0","gas":"0x0","gasUsed":"0x0","to":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d0","input":"0x","value":"0x0","type":"SELFDESTRUCT"}],"value":"0x0","type":"CREATE"}],"value":"0x0","type":"CALL"}}
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
