// The Markdown renderer cannot traverse recursive JSON Schema resources.
// This projection is for display only; validators continue to use openrpc.json.
const fs = require('node:fs');

function forDisplay(value, resource) {
  if (Array.isArray(value)) return value.map(item => forDisplay(item, resource));
  if (value === null || typeof value !== 'object') return value;
  resource = value.$id || resource;
  if (value.$ref === '#') {
    if (!resource) throw new Error('Unscoped recursive schema reference');
    return {
      type: 'object',
      description: `Recursive instance of ${resource}. The child has the same schema as its parent resource; see the canonical OpenRPC schema for validation.`,
    };
  }
  return Object.fromEntries(Object.entries(value).map(([key, child]) => [key, forDisplay(child, resource)]));
}

if (require.main === module) {
  const document = JSON.parse(fs.readFileSync('openrpc.json', 'utf8'));
  fs.writeFileSync('docs-openrpc.json', JSON.stringify(forDisplay(document), null, 2) + '\n');
}
module.exports = {forDisplay};
