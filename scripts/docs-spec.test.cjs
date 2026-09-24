const assert = require('node:assert/strict');
const test = require('node:test');
const {forDisplay} = require('./docs-spec.cjs');

test('recursive resources become labeled display nodes without mutating validation input', () => {
  const input = {$id: 'https://example.org/vm', type: 'object', properties: {sub: {anyOf: [{$ref: '#'}, {type: 'null'}]}}};
  const output = forDisplay(input);
  assert.equal(input.properties.sub.anyOf[0].$ref, '#');
  assert.equal(output.properties.sub.anyOf[0].type, 'object');
  assert.match(output.properties.sub.anyOf[0].description, /https:\/\/example.org\/vm/);
  assert.deepEqual(output.properties.sub.anyOf[1], {type: 'null'});
});
test('unscoped recursion fails and ordinary references are preserved', () => {
  assert.throws(() => forDisplay({$ref: '#'}), /Unscoped/);
  assert.deepEqual(forDisplay({$ref: '#/components/schemas/Value'}), {$ref: '#/components/schemas/Value'});
});

const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const {execFileSync} = require('node:child_process');
const {documentForDisplay} = require('./docs-spec.cjs');

test('actual generated trace pages retain array, variant fields and localization', async () => {
  const {renderMethodsToMarkdown, identityEdits, identitySchemaEdits} = await import('@open-rpc/markdown-generator');
  const input = JSON.parse(fs.readFileSync(path.join(__dirname, '../openrpc.json')));
  const before = JSON.stringify(input);
  const display = documentForDisplay(input);
  display.methods = display.methods.filter(method => method.name.startsWith('trace_'));
  const pages = await renderMethodsToMarkdown(display, identitySchemaEdits, identityEdits);
  for (const page of pages) {
    assert.doesNotMatch(page.markdown, /`unknown(?: or|`)/);
    const method = input.methods.find(method => method.name === page.methodName);
    const parameters = page.markdown.split('## Parameters (by position)')[1].split('## Result')[0];
    for (const param of method.params) {
      const line = parameters.split('\n').find(line => line.startsWith(`**${param.name}** `));
      assert.ok(line, `${method.name} must render parameter ${param.name}`);
      assert.equal(line.includes('*required*'), Boolean(param.required), `${method.name}.${param.name} required status`);
    }
  }
  const many = pages.find(page => page.methodName === 'trace_callMany').markdown;
  assert.match(many, /\*\*Calls\*\* `array<\[object, array<string>\]>` \*required\*/);
  assert.match(many, /\*\*Position 0\*\* `object` \*required\*/);
  assert.match(many, /\*\*Position 1\*\* `array<string>` \*required\*/);
  assert.match(many, /`trace` `stateDiff` `vmTrace`/);
  assert.match(many, /\*\*Items\*\* `\[object, array<string>\]`[\s\S]*?`exactly 2 items`/);
  assert.doesNotMatch(many.split('**Items**')[0], /exactly 2 items/);
  assert.ok(many.indexOf('**Position 0**') < many.indexOf('**Position 1**'));
  assert.ok(many.indexOf('**Calls**') < many.indexOf('**Block**'));
  const example = input.methods.find(method => method.name === 'trace_callMany').examples[0];
  assert.ok(example.params[0].value.length > 0, 'render a nonempty Calls example');
  for (const entry of example.params[0].value) assert.equal(entry.length, 2);
  const block = pages.find(page => page.methodName === 'trace_block').markdown;
  assert.match(block, /\*\*Result\*\* `array<object>/);
  assert.match(block, /different variants may occur in the same array/);
  assert.doesNotMatch(block, /unknown or unknown/);
  for (const field of ['blockHash', 'blockNumber', 'transactionHash', 'transactionPosition', 'action', 'result', 'callType', 'init', 'gasUsed', 'output', 'code']) {
    assert.match(block, new RegExp(`\\*\\*${field}\\*\\* [^\\n]+\\*required\\*`));
  }
  assert.match(block, /Successful CALL/);
  assert.match(block, /Failed CREATE/);
  assert.match(block, /error must be absent/);
  assert.match(block, /error and result are required/);
  assert.match(block, /\*\*transactionHash\*\* `null` \*required\*/);
  const get = pages.find(page => page.methodName === 'trace_get').markdown;
  assert.doesNotMatch(get, /TraceRewardAction/);
  const call = pages.find(page => page.methodName === 'trace_call').markdown;
  assert.match(call, /Recursive instance of \[?https:\/\/ethereum.github.io\/execution-apis\/schemas\/trace-vm.json/);
  assert.doesNotMatch(call, /\*\*blockHash\*\*/);
  for (const field of ['chainId', 'authorizationList', 'blobVersionedHashes', 'maxFeePerBlobGas']) {
    assert.ok(call.includes(`**${field}**`), `standard call field ${field} must be visible`);
  }
  assert.match(call, /Only fields outside this schema are ignored/);
  for (const param of ['StateOverrides', 'BlockOverrides']) assert.ok(call.includes(`**${param}**`), `reserved ${param} parameter must be visible`);
  const filter = pages.find(page => page.methodName === 'trace_filter').markdown;
  for (const field of ['fromAddress', 'toAddress']) {
    assert.ok(filter.includes(`**${field}**`), `nullable filter field ${field} must be visible`);
  }
  assert.match(filter, /Missing, null or empty/);
  assert.match(filter, /-32602/);
  assert.doesNotMatch(filter, /-32001/);
  assert.match(get, /Integer path entries are invalid params/);

  assert.equal(JSON.stringify(input), before);
});

test('projection preserves examples as data', () => {
  const example = {oneOf: [{$ref: '#'}], type: 'array', items: {const: 'value'}};
  const doc = {methods: [{params: [], result: {schema: {type: 'object'}}, examples: [example]}]};
  assert.deepEqual(documentForDisplay(doc).methods[0].examples, [example]);
});

test('watched rebuild refreshes the projection; startup/release hooks preserve prepared specs', () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'docs-refresh-'));
  try {
    const pkg = JSON.parse(fs.readFileSync(path.join(__dirname, '../package.json')));
    assert.match(pkg.scripts.watch, /--exec "npm run docs:refresh"/);
    // Substitute only the expensive Go build with a deterministic source copy.
    // Run the real npm command chain, projection and copy in a scratch site.
    pkg.scripts['build:spec'] = 'node build.cjs';
    fs.mkdirSync(path.join(dir, 'scripts'));
    fs.mkdirSync(path.join(dir, 'docs-api/docs'), {recursive: true});
    fs.writeFileSync(path.join(dir, 'package.json'), JSON.stringify(pkg));
    fs.copyFileSync(path.join(__dirname, 'docs-spec.cjs'), path.join(dir, 'scripts/docs-spec.cjs'));
    fs.writeFileSync(path.join(dir, 'README.md'), 'Watch test');
    fs.writeFileSync(path.join(dir, 'build.cjs'), "require('fs').copyFileSync('source.json','openrpc.json')");
    const run = script => execFileSync('npm', ['run', script], {cwd: dir, stdio: 'pipe'});
    for (const description of ['before edit', 'after edit']) {
      fs.writeFileSync(path.join(dir, 'source.json'), JSON.stringify({methods: [{params: [], result: {schema: {type: 'string', description}}}]}));
      run('docs:refresh');
      const projected = JSON.parse(fs.readFileSync(path.join(dir, 'docs-openrpc.json')));
      assert.equal(projected.methods[0].result.schema.description, description);
    }
    assert.equal(fs.readFileSync(path.join(dir, 'docs-api/docs/quickstart.md'), 'utf8'), 'Watch test');
    // A prepared release must not be rebuilt by either lifecycle hook.
    fs.writeFileSync(path.join(dir, 'build.cjs'), 'process.exit(99)');
    for (const hook of ['prestart', 'prebuild:docusaurus']) {
      fs.rmSync(path.join(dir, 'docs-openrpc.json'));
      run(hook);
      assert.ok(fs.existsSync(path.join(dir, 'docs-openrpc.json')));
    }
  } finally { fs.rmSync(dir, {recursive: true, force: true}); }
});
