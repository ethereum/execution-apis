package testgen

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/program"
)

type haltTraceCase struct {
	name   string
	about  string
	static bool
	gas    uint64
	child  *program.Program
	fault  string
	// reason is part of go-ethereum's error text. It makes the fill fail when
	// the step halts for another reason, such as an ordinary out of gas.
	reason string
}

var haltTraceCases = []haltTraceCase{
	{name: "jump-destination", about: "a JUMP to a byte that is not a JUMPDEST", child: program.New().Push(5).Op(vm.JUMP), fault: "JUMP", reason: "invalid jump destination"},
	{name: "jumpi-destination", about: "a JUMPI with a true condition to a byte that is not a JUMPDEST", child: program.New().Push(1).Push(7).Op(vm.JUMPI), fault: "JUMPI", reason: "invalid jump destination"},
	{name: "opcode-fe", about: "the designated INVALID opcode 0xfe", child: program.New().Op(vm.INVALID), fault: "INVALID", reason: "invalid opcode"},
	{name: "undefined-opcode", about: "the byte 0x0c, which no fork defines", child: program.New().Append([]byte{0x0c}), fault: "opcode 0xc not defined", reason: "invalid opcode"},
	{name: "stack-underflow", about: "an ADD on an empty stack", child: program.New().Op(vm.ADD), fault: "ADD", reason: "stack underflow"},
	{name: "returndata-out-of-bounds", about: "a RETURNDATACOPY of one byte from an empty return data buffer", child: program.New().Push(1).Push(0).Push(0).Op(vm.RETURNDATACOPY), fault: "RETURNDATACOPY", reason: "return data out of bounds"},
	{name: "out-of-gas", about: "a loop that runs out of gas on a PUSH0", gas: 100, child: program.New().Op(vm.JUMPDEST, vm.PUSH0, vm.JUMP), fault: "PUSH0", reason: "out of gas"},
	{name: "memory-out-of-gas", about: "an MLOAD whose memory expansion costs more than the gas left", child: program.New().Push(0xffffffff).Op(vm.MLOAD), fault: "MLOAD", reason: "out of gas"},
	{name: "revert-out-of-gas", about: "a REVERT whose memory expansion costs more than the gas left", child: program.New().Push(0xffffffff).Push(0).Op(vm.REVERT), fault: "REVERT", reason: "out of gas"},
	// The slot is 0 and the SSTORE writes 0, so it costs 2200 and only the
	// sentry, at 2300 gas or less, halts it.
	{name: "sstore-sentry", about: "an SSTORE that does not change its slot, with no more than the 2300 gas of the EIP-2200 sentry", gas: 2300, child: program.New().Push(0).Push(0).Op(vm.SSTORE), fault: "SSTORE", reason: "sentry"},
	{name: "static-log", about: "a LOG0 in a static call", static: true, child: program.New().Push(0).Push(0).Op(vm.LOG0), fault: "LOG0", reason: "write protection"},
	{name: "static-tstore", about: "a TSTORE in a static call", static: true, child: program.New().Push(1).Push(0).Op(vm.TSTORE), fault: "TSTORE", reason: "write protection"},
	{name: "static-call-value", about: "a CALL with value to itself in a static call", static: true, child: program.New().Push(0).Push(0).Push(0).Push(0).Push(1).Push(0xc2).Push(0).Op(vm.CALL), fault: "CALL", reason: "write protection"},
	{name: "static-create", about: "a CREATE in a static call", static: true, child: program.New().Push(0).Push(0).Push(0).Op(vm.CREATE), fault: "CREATE", reason: "write protection"},
	{name: "static-selfdestruct", about: "a SELFDESTRUCT in a static call", static: true, child: program.New().Push(0).Op(vm.SELFDESTRUCT), fault: "SELFDESTRUCT", reason: "write protection"},
}

// haltTraceTests traces, with the opcode logger, a call whose child frame
// halts exceptionally during execution. The parent continues after the child
// fails, so the trace shows that only the child's last step has an error.
func haltTraceTests() []Test {
	var tests []Test
	for _, hc := range haltTraceCases {
		tests = append(tests, hc.test())
	}
	return tests
}

func (hc haltTraceCase) test() Test {
	return Test{
		Name:             "trace-call-halt-" + hc.name,
		About:            fmt.Sprintf("traces a call whose child halts on %s; the failing step is the child's last entry, with an error and a gasCost of 0", hc.about),
		SpecOnly:         true,
		ValidationScript: structLogCompareScript,
		Run: func(ctx context.Context, t *T) error {
			gas := hc.gas
			if gas == 0 {
				// Enough for the constant cost of CREATE and of a CALL with value,
				// so that a static case halts on the static rule.
				gas = 0x10000
			}
			wrapper := program.New()
			if hc.static {
				wrapper.Push(0).Push(0).Push(0).Push(0).Push(0xc2).Push(gas).Op(vm.STATICCALL)
			} else {
				callNoData(wrapper, gas, 0xc2)
			}
			wrapper.Op(vm.POP, vm.STOP)
			sender, _ := t.chain.GetSender(0)
			call := map[string]interface{}{"from": sender, "to": forkAddr(0xc1), "gas": "0x100000"}
			cfg := map[string]interface{}{"stateOverrides": map[string]interface{}{
				forkAddr(0xc1): codeOverride(wrapper),
				forkAddr(0xc2): codeOverride(hc.child),
			}}
			block := hexutil.Uint64(lastBlockBeforeAmsterdam(t.chain).NumberU64())
			var raw json.RawMessage
			if err := t.rpc.CallContext(ctx, &raw, "debug_traceCall", call, block, cfg); err != nil {
				return err
			}
			trace, err := checkFailingSteps(raw, hc.fault+"@2")
			if err != nil {
				return err
			}
			if trace.Failed {
				return fmt.Errorf("expected a successful call: only the child fails")
			}
			for i, log := range trace.StructLogs {
				if log.Error == nil {
					continue
				}
				if !strings.Contains(*log.Error, hc.reason) {
					return fmt.Errorf("structLogs[%d]: error %q, want a halt on %q", i, *log.Error, hc.reason)
				}
				if i+1 >= len(trace.StructLogs) || trace.StructLogs[i+1].Depth != 1 {
					return fmt.Errorf("structLogs[%d]: the failing step is not the child's last entry", i)
				}
			}
			return nil
		},
	}
}
