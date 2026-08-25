// The docs dereferencer fetches absolute-URI $refs over HTTP, but refs that
// anchor recursive schemas resolve in-document via $id and nothing is
// published at those URLs. Stub them out in a website-only copy of the spec.
import fs from 'node:fs';

const doc = JSON.parse(fs.readFileSync('openrpc.json', 'utf8'));

const ids = new Map();
walk(doc, (node) => {
  if (typeof node.$id === 'string') {
    ids.set(node.$id, typeof node.title === 'string' ? node.title : 'schema');
  }
});

walk(doc, (node) => {
  if (typeof node.$ref === 'string' && ids.has(node.$ref)) {
    const title = ids.get(node.$ref);
    delete node.$ref;
    node.title = title;
    node.description =
      `Recursive reference to the "${title}" schema, resolved in-document via its $id.`;
  }
});

fs.writeFileSync('openrpc-docs.json', JSON.stringify(doc, null, 2) + '\n');

function walk(node, visit) {
  if (Array.isArray(node)) {
    for (const item of node) walk(item, visit);
  } else if (node !== null && typeof node === 'object') {
    visit(node);
    for (const value of Object.values(node)) walk(value, visit);
  }
}
