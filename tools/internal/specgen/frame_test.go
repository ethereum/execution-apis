package specgen

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestFrameComponents(t *testing.T) {
	generator := New()
	for _, file := range []string{"base-types.yaml", "transaction.yaml"} {
		content, err := os.ReadFile("../../../src/schemas/" + file)
		if err != nil {
			t.Fatal(err)
		}
		if err := generator.AddSchemas(content); err != nil {
			t.Fatal(err)
		}
	}
	generator.types["FrameTransactionInfo"] = object{
		"allOf": []any{
			object{"$ref": "#/components/schemas/Transaction8141"},
			object{
				"type":       "object",
				"required":   []any{"hash"},
				"properties": object{"hash": object{"$ref": "#/components/schemas/hash32"}},
			},
		},
	}
	frame := `{"mode":"0x1","flags":"0x3","target":null,"executionGas":"0x10000","stateGas":"0x0","value":"0x0","data":"0x"}`
	signature := `{"scheme":"0x0","signer":"0x","msg":"0x","signature":"0xabcd"}`
	envelope := `{"type":"0x6","chainId":"0x1","nonce":"0x0","from":"0x1111111111111111111111111111111111111111","frames":[` + frame + `],"signatures":[],"maxPriorityFeePerGas":"0x1","maxFeePerGas":"0x2","maxFeePerBlobGas":"0x0","blobVersionedHashes":[]}`
	type testCase struct {
		name, schema, input string
		valid               bool
	}
	tests := []testCase{
		{"verify frame", "Frame", frame, true},
		{"default frame", "Frame", strings.ReplaceAll(frame, `"mode":"0x1"`, `"mode":"0x0"`), true},
		{"sender batch", "Frame", strings.ReplaceAll(strings.ReplaceAll(frame, `"mode":"0x1"`, `"mode":"0x2"`), `"flags":"0x3"`, `"flags":"0x4"`), true},
		{"invalid target", "Frame", strings.ReplaceAll(frame, `null`, `"0x1234"`), false},
		{"invalid recovery id", "FrameSignature", `{"scheme":"0x1","signer":"0x","msg":"0x","signature":"0x02` + strings.Repeat("1", 128) + `"}`, false},
		{"explicit cryptographic signer", "FrameSignature", `{"scheme":"0x1","signer":"0x1111111111111111111111111111111111111111","msg":"0x","signature":"0x01` + strings.Repeat("1", 128) + `"}`, true},
		{"odd witness", "FrameSignature", strings.ReplaceAll(signature, "0xabcd", "0xabc"), false},
		{"unknown signature field", "FrameSignature", strings.Replace(signature, `{`, `{"extra":0,`, 1), false},

		{"sender value", "Frame", strings.ReplaceAll(strings.ReplaceAll(frame, `"mode":"0x1"`, `"mode":"0x2"`), `"value":"0x0"`, `"value":"0x1"`), true},
		{"explicit target", "Frame", strings.ReplaceAll(frame, `null`, `"0x1111111111111111111111111111111111111111"`), true},
		{"zero budget", "Frame", strings.ReplaceAll(frame, `"0x10000"`, `"0x0"`), true},
		{"unknown mode", "Frame", strings.ReplaceAll(frame, `"mode":"0x1"`, `"mode":"0x3"`), false},
		{"numeric mode", "Frame", strings.ReplaceAll(frame, `"mode":"0x1"`, `"mode":1`), false},
		{"unknown flag", "Frame", strings.ReplaceAll(frame, `"flags":"0x3"`, `"flags":"0x8"`), false},
		{"batch approval", "Frame", strings.ReplaceAll(frame, `"flags":"0x3"`, `"flags":"0x5"`), false},
		{"missing target", "Frame", strings.ReplaceAll(frame, `"target":null,`, ``), false},
		{"missing budget", "Frame", strings.ReplaceAll(frame, `,"stateGas":"0x0"`, ``), false},
		{"quantity leading zero", "Frame", strings.ReplaceAll(frame, `"0x10000"`, `"0x00"`), false},
		{"unknown frame field", "Frame", strings.ReplaceAll(frame, `"data":"0x"`, `"data":"0x","to":null`), false},
		{"arbitrary witness", "FrameSignature", signature, true},
		{"empty witness", "FrameSignature", strings.ReplaceAll(signature, `0xabcd`, `0x`), true},
		{"explicit digest", "FrameSignature", strings.ReplaceAll(signature, `"msg":"0x"`, `"msg":"0x`+strings.Repeat("1", 64)+`"`), true},
		{"zero digest", "FrameSignature", strings.ReplaceAll(signature, `"msg":"0x"`, `"msg":"0x`+strings.Repeat("0", 64)+`"`), false},
		{"short digest", "FrameSignature", strings.ReplaceAll(signature, `"msg":"0x"`, `"msg":"0x01"`), false},
		{"arbitrary signer", "FrameSignature", strings.ReplaceAll(signature, `"signer":"0x"`, `"signer":"0x`+strings.Repeat("1", 40)+`"`), false},
		{"null signer", "FrameSignature", strings.ReplaceAll(signature, `"signer":"0x"`, `"signer":null`), false},
		{"unknown scheme", "FrameSignature", strings.ReplaceAll(signature, `"scheme":"0x0"`, `"scheme":"0x3"`), false},
		{"complete envelope", "Transaction8141", envelope, true},
		{"composed lookup metadata", "FrameTransactionInfo", strings.Replace(envelope, `{`, `{"hash":"0x`+strings.Repeat("1", 64)+`",`, 1), true},
		{"missing lookup metadata", "FrameTransactionInfo", envelope, false},
		{"lookup metadata", "Transaction8141", strings.Replace(envelope, `{`, `{"hash":"0x`+strings.Repeat("1", 64)+`","blockNumber":"0x1",`, 1), true},
		{"missing from", "Transaction8141", strings.ReplaceAll(envelope, `"from":`, `"sender":`), false},
		{"nested limits", "Frame", strings.ReplaceAll(frame, `"executionGas":"0x10000","stateGas":"0x0"`, `"limits":{"execution":"0x10000","state":"0x0"}`), false},
		{"malformed execution gas", "Frame", strings.ReplaceAll(frame, `"executionGas":"0x10000"`, `"executionGas":"10000"`), false},
		{"malformed state gas", "Frame", strings.ReplaceAll(frame, `"stateGas":"0x0"`, `"stateGas":0`), false},
		{"malformed calldata", "Frame", strings.ReplaceAll(frame, `"data":"0x"`, `"data":"0xzz"`), false},
		{"short blob hash", "Transaction8141", strings.ReplaceAll(envelope, `"blobVersionedHashes":[]`, `"blobVersionedHashes":["0x01"]`), false},
		{"nested fees", "Transaction8141", strings.ReplaceAll(envelope, `"maxPriorityFeePerGas":"0x1","maxFeePerGas":"0x2","maxFeePerBlobGas":"0x0"`, `"fees":{"maxPriorityFeePerGas":"0x1","maxFeePerGas":"0x2","maxFeePerBlobGas":"0x0"}`), false},
		{"64 frames", "Transaction8141", strings.ReplaceAll(envelope, frame, strings.TrimSuffix(strings.Repeat(frame+",", 64), ",")), true},
		{"byte type", "Transaction8141", strings.ReplaceAll(envelope, `"type":"0x6"`, `"type":"0x06"`), false},
		{"blob envelope", "Transaction8141", strings.ReplaceAll(strings.ReplaceAll(envelope, `"blobVersionedHashes":[]`, `"blobVersionedHashes":["0x01`+strings.Repeat("0", 62)+`"]`), `"maxFeePerBlobGas":"0x0"`, `"maxFeePerBlobGas":"0x1"`), true},
	}
	for _, field := range []string{"chainId", "maxFeePerGas", "maxPriorityFeePerGas", "maxFeePerBlobGas"} {
		var input map[string]any
		if err := json.Unmarshal([]byte(envelope), &input); err != nil {
			t.Fatal(err)
		}
		input["blobVersionedHashes"] = []string{"0x01" + strings.Repeat("0", 62)}
		for _, width := range []int{64, 65} {
			input[field] = "0x" + strings.Repeat("f", width)
			data, err := json.Marshal(input)
			if err != nil {
				t.Fatal(err)
			}
			tests = append(tests, testCase{field + " " + strconv.Itoa(width) + " hex digits", "Transaction8141", string(data), true})
		}
	}
	for _, scheme := range []struct {
		id     string
		size   int
		prefix string
	}{{"0x1", 128, "00"}, {"0x2", 256, ""}} {
		input := strings.ReplaceAll(signature, `"scheme":"0x0"`, `"scheme":"`+scheme.id+`"`)
		for _, variant := range []struct {
			name, raw string
			valid     bool
		}{
			{"complete", "0x" + scheme.prefix + strings.Repeat("1", scheme.size), true},
			{"empty", "0x", false},
			{"short", "0x" + scheme.prefix + strings.Repeat("1", scheme.size-2), false},
			{"long", "0x" + scheme.prefix + strings.Repeat("1", scheme.size+2), false},
		} {
			entry := strings.ReplaceAll(input, "0xabcd", variant.raw)
			tests = append(tests, testCase{"signature " + scheme.id + " " + variant.name, "FrameSignature", entry, variant.valid})
			tests = append(tests, testCase{"envelope signature " + scheme.id + " " + variant.name, "Transaction8141", strings.ReplaceAll(envelope, `"signatures":[]`, `"signatures":[`+entry+`]`), variant.valid})
		}

	}
	for _, tc := range []struct {
		name, input string
		valid       bool
	}{
		{"secp256k1", `{"scheme":"0x1","signer":"0x","msg":"0x"}`, true},
		{"p256", `{"scheme":"0x2","signer":"0x1111111111111111111111111111111111111111","msg":"0x"}`, true},
		{"arbitrary", `{"scheme":"0x0","signer":"0x","msg":"0x"}`, true},
		{"arbitrary signer", `{"scheme":"0x0","signer":"0x1111111111111111111111111111111111111111","msg":"0x"}`, false},
		{"unknown scheme", `{"scheme":"0x3","signer":"0x","msg":"0x"}`, false},
		{"nonempty bytes", `{"scheme":"0x1","signer":"0x","msg":"0x","signature":"0x01"}`, false},
		{"null bytes", `{"scheme":"0x1","signer":"0x","msg":"0x","signature":null}`, false},
		{"short signer", `{"scheme":"0x1","signer":"0x01","msg":"0x"}`, false},
		{"zero digest", `{"scheme":"0x1","signer":"0x","msg":"0x` + strings.Repeat("0", 64) + `"}`, false},
		{"explicit digest", `{"scheme":"0x1","signer":"0x","msg":"0x` + strings.Repeat("1", 64) + `"}`, true},
		{"empty bytes", `{"scheme":"0x1","signer":"0x","msg":"0x","signature":"0x"}`, false},
	} {
		tests = append(tests, testCase{"placeholder " + tc.name, "FrameSignaturePlaceholder", tc.input, tc.valid})
		tests = append(tests, testCase{"envelope placeholder " + tc.name, "Transaction8141", strings.ReplaceAll(envelope, `"signatures":[]`, `"signatures":[`+tc.input+`]`), tc.valid})
	}
	for _, component := range []struct{ name, input string }{{"Frame", frame}, {"FrameSignature", signature}, {"FrameSignaturePlaceholder", strings.ReplaceAll(signature, `,"signature":"0xabcd"`, "")}, {"Transaction8141", envelope}} {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal([]byte(component.input), &fields); err != nil {
			t.Fatal(err)
		}
		for field, value := range fields {
			delete(fields, field)
			input, err := json.Marshal(fields)
			if err != nil {
				t.Fatal(err)
			}
			tests = append(tests, testCase{component.name + " missing " + field, component.name, string(input), false})
			fields[field] = value
		}
	}

	for _, expanded := range []bool{false, true} {
		for _, tc := range tests {
			mode := "referenced/"
			if expanded {
				mode = "expanded/"
			}
			t.Run(mode+tc.name, func(t *testing.T) {
				schema := generator.types[tc.schema]
				if expanded {
					var err error
					schema, err = generator.expandSchema(schema, generator.types)
					if err != nil {
						t.Fatal(err)
					}
				}
				compiler := jsonschema.NewCompiler()
				compiler.DefaultDraft(jsonschema.Draft7)
				root := object{"components": object{"schemas": repo2object(generator.types)}, "$ref": "#/components/schemas/" + tc.schema}
				if expanded {
					root = schema
				}
				if err := compiler.AddResource("frame.json", root); err != nil {
					t.Fatal(err)
				}
				compiled, err := compiler.Compile("frame.json")
				if err != nil {
					t.Fatal(err)
				}
				var input any
				if err := json.Unmarshal([]byte(tc.input), &input); err != nil {
					t.Fatal(err)
				}
				err = compiled.Validate(input)
				if (err == nil) != tc.valid {
					t.Fatalf("valid=%v, validation error: %v", tc.valid, err)
				}
			})
		}
	}
}
