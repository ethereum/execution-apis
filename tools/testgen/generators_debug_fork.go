package testgen

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/program"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

var forkActive = map[string]func(*params.ChainConfig, *types.Header) bool{
	"Byzantium":      func(c *params.ChainConfig, h *types.Header) bool { return c.IsByzantium(h.Number) },
	"Constantinople": func(c *params.ChainConfig, h *types.Header) bool { return c.IsConstantinople(h.Number) },
	"Petersburg":     func(c *params.ChainConfig, h *types.Header) bool { return c.IsPetersburg(h.Number) },
	"Istanbul":       func(c *params.ChainConfig, h *types.Header) bool { return c.IsIstanbul(h.Number) },
	"Berlin":         func(c *params.ChainConfig, h *types.Header) bool { return c.IsBerlin(h.Number) },
	"London":         func(c *params.ChainConfig, h *types.Header) bool { return c.IsLondon(h.Number) },
	"Paris":          func(_ *params.ChainConfig, h *types.Header) bool { return h.Difficulty.Sign() == 0 },
	"Shanghai":       func(c *params.ChainConfig, h *types.Header) bool { return c.IsShanghai(h.Number, h.Time) },
	"Cancun":         func(c *params.ChainConfig, h *types.Header) bool { return c.IsCancun(h.Number, h.Time) },
	"Prague":         func(c *params.ChainConfig, h *types.Header) bool { return c.IsPrague(h.Number, h.Time) },
	"Osaka":          func(c *params.ChainConfig, h *types.Header) bool { return c.IsOsaka(h.Number, h.Time) },
	"Amsterdam":      func(c *params.ChainConfig, h *types.Header) bool { return c.IsAmsterdam(h.Number, h.Time) },
}

func firstForkBlock(c *Chain, fork string) uint64 {
	for _, b := range c.blocks {
		if forkActive[fork](c.Config(), b.Header()) {
			return b.NumberU64()
		}
	}
	panic("fork " + fork + " is not active in the chain")
}

func forkAddr(n uint64) string {
	return fmt.Sprintf("0x%040x", n)
}

func codeOverride(p *program.Program) map[string]interface{} {
	return map[string]interface{}{"code": hexutil.Encode(p.Bytes())}
}

// program.Call replaces all-zero arguments with DUP1, which changes the
// recorded traces.
func callNoData(p *program.Program, gas, addr uint64) *program.Program {
	return p.Push(0).Push(0).Push(0).Push(0).Push(0).Push(addr).Push(gas).Op(vm.CALL)
}

// withCodeData copies data, which follows the code, to memory offset 0 and
// then runs body. The data offset is a PUSH2, so its size does not change the
// code before it.
func withCodeData(data []byte, body func(*program.Program)) *program.Program {
	p := program.New().Push(len(data)).Op(vm.PUSH2)
	pos := p.Size()
	p.Append([]byte{0, 0}).Push(0).Op(vm.CODECOPY)
	body(p)
	offset := p.Size()
	p.Append(data)
	p.Bytes()[pos], p.Bytes()[pos+1] = byte(offset>>8), byte(offset)
	return p
}

// Each opcode runs in its own child, so an undefined opcode fails only its
// own frame.
func forkChildren(children ...*program.Program) map[string]interface{} {
	w := program.New()
	ov := map[string]interface{}{}
	for i, child := range children {
		addr := 0xd1 + uint64(i)
		callNoData(w, 0x4000, addr).Op(vm.POP)
		ov[forkAddr(addr)] = codeOverride(child)
	}
	ov[forkAddr(0xc1)] = codeOverride(w.Op(vm.STOP))
	return ov
}

func ops(codes ...vm.OpCode) *program.Program {
	return program.New().Op(codes...)
}

type forkSide struct {
	at     bool
	masked bool
}

var (
	forkPre       = forkSide{}
	forkPreMasked = forkSide{masked: true}
	forkAt        = forkSide{at: true}
	forkAtMasked  = forkSide{at: true, masked: true}
)

type forkTraceCase struct {
	fork         string
	name         string
	about        string
	to           string
	input        string
	overrides    map[string]interface{}
	withLog      bool
	withFee      bool
	opcodeConfig map[string]interface{}
	callTracer   []forkSide
	structLog    []forkSide
	// atFork moves the "at" side to the first block of a later fork. The test
	// chain can activate Constantinople before Petersburg. No live
	// network has blocks between the two forks, and the execution specs define
	// one Constantinople fork without EIP-1283, so clients may run those blocks
	// under other rules. Do not target a block between the two forks.
	atFork string
}

var forkTraceCases = func() []forkTraceCase {
	var (
		wrapper     = forkAddr(0xc1)
		target      = forkAddr(0xc2)
		delegate    = forkAddr(0xc3)
		delegator   = forkAddr(0xe7)
		storageAcct = "0x8bebc8ba651aee624937e7d897853ac30c95a067"
		gas         = uint256.NewInt(0x10000)

		errorBoom = hexutil.MustDecode("0x08c379a0" + fmt.Sprintf("%064x%064x", 0x20, 4) + fmt.Sprintf("%x", "boom") + strings.Repeat("0", 56))
		revert    = withCodeData(errorBoom, func(p *program.Program) { p.Push(100).Push(0).Op(vm.REVERT) })
		return1   = program.New().Push(1).Push(0).Op(vm.MSTORE).Return(0, 32)
		init2a    = program.New().Push(0x2a).Push(0).Op(vm.MSTORE8).Return(0, 1).Bytes()
		initEF    = program.New().Push(0xef).Push(0).Op(vm.MSTORE8).Return(0, 1).Bytes()
		// SSTORE 1->0, 2->2, 3->7->3 and an SLOAD of slot 1, on slots that
		// hold 1, 2 and 3 in the genesis block.
		sstores = program.New().Sstore(1, 0).Sstore(2, 2).Sstore(3, 7).Sstore(3, 3).Push(1).Op(vm.SLOAD, vm.POP, vm.STOP)
		g1      = hexutil.MustDecode("0x" + strings.Repeat("0", 32) + "17f1d3a73197d7942695638c4fa9ac0fc3688c4f9774b905a14e3a3f171bac586c55e83ff97a1aeffb3af00adb22c6bb" +
			strings.Repeat("0", 32) + "08b3f481e3aaa0f1a09e30ed741d8ae4fcf5e095d5d00af600db18cb2c04b3edd03cc744a2888ae40caa232946c5e7e1")
		p256Input   = hexutil.MustDecode("0xfe63c5953b939ea4e464a797d55866925155e3331c33be5441afb99ed51382b9e980a158df91a22d613e844e8ff3153ea5191d7226b73c9a1f4942de942fbf496f98a258208217c9ecb0315328b0289a23e31d5db6b56054b8b9ec23e5ebfe589fad84aeae08bbef7f010014d82cef6a09de2b0cf871b5ce0c4f1d13a59a593407cb45769f1070e2c2470fe5b1bfe63133c0b0cdc64ea4bf3791a8ec2a07fd4f")
		modexpInput = hexutil.MustDecode("0x" + strings.Repeat(fmt.Sprintf("%064x", 1), 3) + "020305")
		delegation  = map[string]interface{}{"code": "0xef0100" + delegate[2:]}
		// Emits a LOG1 with topic 0xaa and returns its own address.
		delegateCode = program.New().Op(vm.ADDRESS).Push(0).Op(vm.MSTORE).Push(0xaa).Push(32).Push(0).Op(vm.LOG1).Return(0, 32)
		callTarget   = callNoData(program.New(), 0x10000, 0xc2).Op(vm.POP, vm.STOP)
	)
	selfdestruct := map[string]interface{}{
		wrapper: codeOverride(callNoData(program.New(), 0x10000, 0xc2).Op(vm.POP).
			Push(0xc2).Op(vm.EXTCODESIZE).Push(0xc2).Op(vm.BALANCE).Push(0xbe).Op(vm.BALANCE, vm.STOP)),
		target: map[string]interface{}{"code": hexutil.Encode(program.New().Selfdestruct(0xbe).Bytes()), "balance": "0xde0b6b3a7640000"},
	}
	both := []forkSide{forkPre, forkAt}
	newOpcode := []forkSide{forkPreMasked, forkAt}
	sstore := func(fork string) forkTraceCase {
		return forkTraceCase{
			fork: fork, name: "sstore-refund", about: "SSTORE gas and refunds on slots with committed values",
			to: storageAcct, overrides: map[string]interface{}{storageAcct: codeOverride(sstores)}, callTracer: both,
		}
	}
	return []forkTraceCase{
		{
			fork: "Byzantium", name: "revert", about: "a child call that executes REVERT with Error(string) data",
			to: wrapper, overrides: map[string]interface{}{wrapper: codeOverride(callTarget), target: codeOverride(revert)},
			callTracer: newOpcode,
		},
		{
			fork: "Byzantium", name: "revert-root", about: "a call that executes REVERT with Error(string) data",
			to: target, overrides: map[string]interface{}{target: codeOverride(revert)}, callTracer: newOpcode,
		},
		{
			fork: "Byzantium", name: "staticcall", about: "a STATICCALL to a contract that returns one word",
			to: wrapper, overrides: map[string]interface{}{
				wrapper: codeOverride(program.New().StaticCall(gas, 0xc2, 0, 0, 0, 32).Op(vm.POP, vm.STOP)),
				target:  codeOverride(return1),
			},
			callTracer: newOpcode, structLog: []forkSide{forkAt},
		},
		{
			fork: "Byzantium", name: "returndata", about: "RETURNDATASIZE and RETURNDATACOPY after a child call",
			to: wrapper, overrides: map[string]interface{}{
				wrapper: codeOverride(callNoData(program.New(), 0x10000, 0xc2).Op(vm.POP, vm.RETURNDATASIZE, vm.POP).
					Push(32).Push(0).Push(0).Op(vm.RETURNDATACOPY, vm.STOP)),
				target: codeOverride(return1),
			},
			opcodeConfig: map[string]interface{}{"enableReturnData": true},
			callTracer:   newOpcode, structLog: []forkSide{forkAt},
		},
		{
			fork: "Constantinople", atFork: "Petersburg", name: "create2", about: "a CREATE2 that deploys one byte",
			to: wrapper, overrides: map[string]interface{}{wrapper: codeOverride(program.New().MstoreSmall(init2a, 0).
				Push(1).Push(len(init2a)).Push(32-len(init2a)).Push(0).Op(vm.CREATE2, vm.POP, vm.STOP))},
			callTracer: newOpcode, structLog: []forkSide{forkAt},
		},
		{
			fork: "Constantinople", atFork: "Petersburg", name: "shift-extcodehash", about: "SHL, SHR, SAR and EXTCODEHASH, each in its own child call",
			to: wrapper, overrides: forkChildren(
				program.New().Push(1).Push(1).Op(vm.SHL, vm.POP, vm.STOP),
				program.New().Push(1).Push(1).Op(vm.SHR, vm.POP, vm.STOP),
				program.New().Push(1).Push(1).Op(vm.SAR, vm.POP, vm.STOP),
				ops(vm.ADDRESS, vm.EXTCODEHASH, vm.POP, vm.STOP),
			),
			callTracer: newOpcode, structLog: []forkSide{forkAt},
		},
		sstore("Istanbul"),
		sstore("London"),
		{
			fork: "Istanbul", name: "chainid-selfbalance", about: "CHAINID and SELFBALANCE, each in its own child call",
			to: wrapper, overrides: forkChildren(ops(vm.CHAINID, vm.POP, vm.STOP), ops(vm.SELFBALANCE, vm.POP, vm.STOP)),
			callTracer: newOpcode, structLog: []forkSide{forkAt},
		},
		{
			fork: "Berlin", name: "cold-access", about: "two SLOADs of one slot and two calls to one account",
			to: wrapper, overrides: map[string]interface{}{
				wrapper: codeOverride(callNoData(callNoData(program.New().Push(5).Op(vm.SLOAD, vm.POP).Push(5).Op(vm.SLOAD, vm.POP),
					0x1000, 0xc2).Op(vm.POP), 0x1000, 0xc2).Op(vm.POP, vm.STOP)),
				target: codeOverride(program.New().Push(1).Op(vm.POP, vm.STOP)),
			},
			callTracer: both, structLog: both,
		},
		{
			fork: "London", name: "basefee", about: "a call with a gas price and a funded sender, whose child call executes BASEFEE",
			to: wrapper, overrides: forkChildren(ops(vm.BASEFEE, vm.POP, vm.STOP)), withFee: true,
			callTracer: newOpcode, structLog: []forkSide{forkAt},
		},
		{
			fork: "London", name: "ef-code-create", about: "a CREATE whose initcode returns code that starts with 0xef",
			to: wrapper, overrides: map[string]interface{}{wrapper: codeOverride(program.New().MstoreSmall(initEF, 0).
				Push(len(initEF)).Push(32-len(initEF)).Push(0).Op(vm.CREATE, vm.POP, vm.STOP))},
			callTracer: []forkSide{forkPre, forkAtMasked}, structLog: []forkSide{forkPre},
		},
		{
			fork: "London", name: "ef-code-create-root", about: "a contract creation whose initcode returns code that starts with 0xef",
			input:      hexutil.Encode(initEF),
			callTracer: []forkSide{forkPre, forkAtMasked}, structLog: []forkSide{forkPre},
		},
		{
			fork: "London", name: "selfdestruct-refund", about: "a child call that executes SELFDESTRUCT",
			to: wrapper, overrides: selfdestruct,
			callTracer: both, structLog: []forkSide{forkAt},
		},
		{
			fork: "Paris", name: "difficulty", about: "a call that returns the DIFFICULTY (PREVRANDAO) value",
			to: target, overrides: map[string]interface{}{target: codeOverride(program.New().Op(vm.DIFFICULTY).Push(0).Op(vm.MSTORE).Return(0, 32))},
			callTracer: both, structLog: []forkSide{forkPre},
		},
		{
			fork: "Shanghai", name: "push0", about: "PUSH0 in a child call",
			to: wrapper, overrides: forkChildren(ops(vm.PUSH0, vm.POP, vm.STOP)),
			callTracer: newOpcode, structLog: []forkSide{forkAt},
		},
		{
			fork: "Shanghai", name: "initcode-limit", about: "a child call that executes CREATE with 49153 bytes of initcode",
			to: wrapper, overrides: map[string]interface{}{
				wrapper: codeOverride(callNoData(program.New(), 0x20000, 0xc2).Op(vm.POP, vm.STOP)),
				target:  codeOverride(program.New().Push(params.MaxInitCodeSize+1).Push(0).Push(0).Op(vm.CREATE, vm.POP, vm.STOP)),
			},
			callTracer: []forkSide{forkPre, forkAtMasked}, structLog: []forkSide{forkPre},
		},
		{
			fork: "Shanghai", name: "warm-coinbase", about: "a call that reads the BALANCE of the COINBASE account",
			to: target, overrides: map[string]interface{}{target: codeOverride(ops(vm.COINBASE, vm.BALANCE, vm.POP, vm.STOP))},
			callTracer: both, structLog: both,
		},
		{
			fork: "Cancun", name: "opcodes", about: "TSTORE, TLOAD, MCOPY, BLOBHASH and BLOBBASEFEE, each in its own child call",
			to: wrapper, overrides: forkChildren(
				program.New().Tstore(0, 1).Op(vm.STOP),
				program.New().Push(0).Op(vm.TLOAD, vm.POP, vm.STOP),
				program.New().Push(1).Push(0).Push(0).Op(vm.MCOPY, vm.STOP),
				program.New().Push(0).Op(vm.BLOBHASH, vm.POP, vm.STOP),
				ops(vm.BLOBBASEFEE, vm.POP, vm.STOP),
			),
			callTracer: newOpcode, structLog: []forkSide{forkAt},
		},
		{
			fork: "Prague", name: "delegation", about: "a child call to an account with an EIP-7702 delegation designator",
			to: wrapper, overrides: map[string]interface{}{
				wrapper:   codeOverride(program.New().Call(gas, 0xe7, 0, 0, 0, 0, 32).Op(vm.POP, vm.STOP)),
				delegator: delegation,
				delegate:  codeOverride(delegateCode),
			},
			withLog:    true,
			callTracer: newOpcode, structLog: []forkSide{forkAt},
		},
		{
			fork: "Prague", name: "delegation-root", about: "a call to an account with an EIP-7702 delegation designator",
			to: delegator, overrides: map[string]interface{}{delegator: delegation, delegate: codeOverride(delegateCode)},
			withLog:    true,
			callTracer: newOpcode, structLog: []forkSide{forkAt},
		},
		{
			fork: "Prague", name: "bls12-g1add", about: "a call to the BLS12_G1ADD precompile address",
			to: wrapper, overrides: map[string]interface{}{wrapper: codeOverride(withCodeData(append(g1, g1...), func(p *program.Program) {
				p.Call(gas, 0x0b, 0, 0, 256, 256, 128).Op(vm.POP, vm.STOP)
			}))},
			callTracer: both, structLog: both,
		},
		{
			fork: "Osaka", name: "clz", about: "CLZ in a child call",
			to: wrapper, overrides: forkChildren(program.New().Push(1).Op(vm.CLZ, vm.POP, vm.STOP)),
			callTracer: newOpcode, structLog: []forkSide{forkAt},
		},
		{
			fork: "Osaka", name: "p256verify", about: "a call to the P256VERIFY precompile address with a valid signature",
			to: wrapper, overrides: map[string]interface{}{wrapper: codeOverride(withCodeData(p256Input, func(p *program.Program) {
				p.Call(gas, 0x100, 0, 0, 160, 160, 32).Op(vm.POP, vm.STOP)
			}))},
			callTracer: both, structLog: both,
		},
		{
			fork: "Osaka", name: "modexp", about: "a call to the MODEXP precompile with one-byte operands",
			to: wrapper, overrides: map[string]interface{}{wrapper: codeOverride(withCodeData(modexpInput, func(p *program.Program) {
				p.Call(gas, 0x05, 0, 0, 99, 99, 1).Op(vm.POP, vm.STOP)
			}))},
			callTracer: both, structLog: both,
		},
	}
}()

func forkTraceTests() []Test {
	var tests []Test
	for _, fc := range forkTraceCases {
		for _, side := range fc.callTracer {
			tests = append(tests, fc.test(true, side))
		}
		for _, side := range fc.structLog {
			tests = append(tests, fc.test(false, side))
		}
	}
	return tests
}

func (fc forkTraceCase) test(callTracer bool, side forkSide) Test {
	prefix, tracerName := "trace-call", "the default struct logger"
	if callTracer {
		prefix, tracerName = "calltracer", "the callTracer"
	}
	atFork := fc.fork
	if fc.atFork != "" {
		atFork = fc.atFork
	}
	sideName, where := "pre", "the last block before "+fc.fork
	if side.at {
		sideName, where = "at", "the first "+atFork+" block"
	}
	test := Test{
		Name:     fmt.Sprintf("%s-fork-%s-%s-%s", prefix, strings.ToLower(fc.fork), fc.name, sideName),
		About:    fmt.Sprintf("traces %s with %s, at %s", fc.about, tracerName, where),
		SpecOnly: side.masked,
	}
	if side.masked {
		test.ValidationScript = callFrameCompareScript
	}
	test.Run = func(ctx context.Context, t *T) error {
		sender, _ := t.chain.GetSender(0)
		var block uint64
		if side.at {
			block = firstForkBlock(t.chain, atFork)
		} else {
			block = firstForkBlock(t.chain, fc.fork) - 1
		}
		call := map[string]interface{}{"from": sender, "gas": "0x100000"}
		if fc.to != "" {
			call["to"] = fc.to
		}
		if fc.input != "" {
			call["input"] = fc.input
		}
		overrides := fc.overrides
		if fc.withFee {
			// The London base fee starts at InitialBaseFee, so this price covers
			// it and BASEFEE reads the block's value instead of a zero-fee call's.
			fee := "gasPrice"
			if forkActive["London"](t.chain.Config(), t.chain.GetBlock(int(block)).Header()) {
				fee = "maxFeePerGas"
			}
			call[fee] = hexutil.EncodeUint64(params.InitialBaseFee)
			overrides = map[string]interface{}{strings.ToLower(sender.Hex()): map[string]interface{}{"balance": "0xde0b6b3a7640000"}}
			for k, v := range fc.overrides {
				overrides[k] = v
			}
		}
		cfg := map[string]interface{}{}
		if callTracer {
			cfg["tracer"] = "callTracer"
			if fc.withLog {
				cfg["tracerConfig"] = map[string]interface{}{"withLog": true}
			}
		} else {
			for k, v := range fc.opcodeConfig {
				cfg[k] = v
			}
		}
		if overrides != nil {
			cfg["stateOverrides"] = overrides
		}
		var result map[string]interface{}
		if err := t.rpc.CallContext(ctx, &result, "debug_traceCall", call, hexutil.Uint64(block), cfg); err != nil {
			return err
		}
		if callTracer {
			return validateCallFrame(result, callTracerOpts{withLog: fc.withLog})
		}
		return validateOpcodeTransactionTrace(result)
	}
	return test
}
