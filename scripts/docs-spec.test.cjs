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
