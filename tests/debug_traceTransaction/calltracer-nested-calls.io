// traces a calltree contract invocation covering all call frame types: nested calls, inner revert with reason, staticcall write-protection failure, delegatecall, callcode, precompile call, create and selfdestruct
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"debug_traceTransaction","params":["0xa3cec99b572cbb16af9a680a23b7925523b286df43c56548f52f3b1677c99189",{"tracer":"callTracer"}]}
<< {"jsonrpc":"2.0","id":1,"result":{"from":"0x7435ed30a8b4aeb0877cef0c6e8cffe834eb865f","gas":"0x927c0","gasUsed":"0x28645","to":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d0","input":"0x","output":"0xffee","calls":[{"from":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d0","gas":"0xf35c","gasUsed":"0x48","to":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d1","input":"0xff01","output":"0xffee","value":"0x1","type":"CALL"},{"from":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d0","gas":"0xea60","gasUsed":"0x889","to":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d3","input":"0x0000000000000000000000000000000000000000000000000000000000000001","output":"0x08c379a00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000a75736572206572726f72","error":"execution reverted","revertReason":"user error","value":"0x0","type":"CALL"},{"from":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d0","gas":"0xea60","gasUsed":"0x7d","to":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d2","input":"0x","output":"0x000000000000000000000000000000000000000000000000000000000000001b000000000000000000000000000000000000000000000000000c72dd9d5e883e0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003b9aca0000000000000000000000000000000000000000000000000000000000000202c00000000000000000000000007435ed30a8b4aeb0877cef0c6e8cffe834eb865f0000000000000000000000000000000000000000000000000000000000000000","type":"STATICCALL"},{"from":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d0","gas":"0xea60","gasUsed":"0xea60","to":"0x7dcd17433742f4c0ca53122ab541d0ba67fc27df","input":"0x","error":"out of gas: write protection","type":"STATICCALL"},{"from":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d0","gas":"0xea60","gasUsed":"0x64b0","to":"0x7dcd17433742f4c0ca53122ab541d0ba67fc27df","input":"0x00000000","value":"0x0","type":"DELEGATECALL"},{"from":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d0","gas":"0xea60","gasUsed":"0x48","to":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d1","input":"0xff01","output":"0xffee","value":"0x0","type":"CALLCODE"},{"from":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d0","gas":"0xea60","gasUsed":"0x12","to":"0x0000000000000000000000000000000000000004","input":"0xff010000","output":"0xff010000","value":"0x0","type":"CALL"},{"from":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d0","gas":"0x6a011","gasUsed":"0x1682","to":"0x2d303c5b7911d87d594bf1b31fbb9aa187888893","input":"0x5b646368696c6460006000a133ff","calls":[{"from":"0x2d303c5b7911d87d594bf1b31fbb9aa187888893","gas":"0x0","gasUsed":"0x0","to":"0x9dcd17433742f4c0ca53122ab541d0ba67fc27d0","input":"0x","value":"0x0","type":"SELFDESTRUCT"}],"value":"0x0","type":"CREATE"}],"value":"0x0","type":"CALL"}}
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
