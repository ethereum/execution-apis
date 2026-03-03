# Tools

This directory contains the tooling for building, validating, and testing the
Execution API specification. The tools are used by CI and by contributors
preparing PRs.

For the contribution workflow and how these tools fit in, see
[CONTRIBUTING.md](../CONTRIBUTING.md) and the
[Contributors Guide](../docs-api/docs/contributors-guide.md).

## Overview

| Tool       | Purpose                                              |
| ---------- | ---------------------------------------------------- |
| specgen    | Compiles YAML spec files into `openrpc.json`         |
| speccheck  | Validates test fixtures in `tests/` against the spec |
| rpctestgen | Generates `.io` fixtures by running tests vs geth    |

## Passing CI

The CI pipeline runs the following. To pass locally before opening a PR:

1. **Build the spec** — `make build` (from repo root)
2. **Run speccheck** — `make test` (from repo root)
3. **Fill tests** — `make fill` (from repo root, or `make fill` from `tools/`)
4. **Lint tools** — `make lint` (from `tools/`)

If any of these fail, CI will fail. For new methods that require upstream
go-ethereum changes, rpctestgen may not pass until go-ethereum implements
them; see [CONTRIBUTING.md](../CONTRIBUTING.md) for CI exception policy.

### Build

From the repo root:

```console
$ make build
...
wrote spec to openrpc.json
```

This builds the tools and runs specgen to produce `openrpc.json` and
`refs-openrpc.json`.

### speccheck

Validates that test cases in `tests/` conform to the OpenRPC specification.
**Must pass** for all PRs.

From the repo root:

```console
$ make test
...
all passing.
```

Or from `tools/` after building:

```console
$ ./speccheck -v
all passing.
```

Options: `--spec` (default: `openrpc.json`), `--tests` (default: `tests`),
`--regexp` (filter tests), `-v` (verbose).

### rpctestgen (fill)

Generates test fixtures by executing tests against a geth client. Uses
go-ethereum libraries; requires geth to be built. The `make fill` target builds
geth and rpctestgen, then runs the generator.

From the repo root or `tools/`:

```console
$ make fill
...
generating tests/eth_blockNumber/simple-test.io  done.
```

Output is written to `tests/`. CI checks that no files change after `make fill`
(except known non-deterministic tests). For new methods not yet in go-ethereum,
this step may fail; maintainers may merge with CI exceptions.

Options: `--bin` (client binary), `--chain` (chain dir), `--out` (output dir),
`--tests` (regex filter), `-v` (verbose). Run `./rpctestgen --help` for details.

### Lint

From `tools/`:

```console
$ make lint
(no output when all checks pass)
```

Runs `gofmt`, `go vet`, and `staticcheck`.
Install staticcheck with `go install honnef.co/go/tools/cmd/staticcheck@latest`
if needed.

## specgen

Compiles the YAML method and schema files into a single OpenRPC document. Used
internally by `make build`; typically not run directly by contributors.

## speccheck (details)

Validates test fixtures against the spec. See [speccheck](#speccheck) above.

## rpctestgen (details)

Test fixture generator. Runs test definitions against a client (default: geth)
and records the request-response exchange. See [rpctestgen (fill)](#rpctestgen-fill)
above.

### Fixture format

Fixtures use a simple line-delimited format. `>>` denotes a request;
`<<` denotes the response.

```javascript
>> {"jsonrpc":"2.0","id":1,"method":"eth_blockNumber"}
<< {"jsonrpc":"2.0","id":1,"result":"0x3"}
```

Tests are stored at `tests/{method-name}/{test-name}.io`. The generator also
outputs `chain.rlp` and `genesis.json` so exchanges can be verified on all
clients.

For more on test format and chain making, see the
[Tests documentation](../docs-api/docs/tests.md).

## Documentation

When you change the spec or docs, rebuild the documentation site:

```console
$ npm run build:docs
...
```

Commit any generated doc updates. The spell checker (`pyspelling -c
spellcheck.yaml`) runs in CI; fix spelling errors before pushing. When adding
new method names, you may need to add them (or their parts) to
[wordlist.txt](../wordlist.txt) so the spell checker accepts them.
