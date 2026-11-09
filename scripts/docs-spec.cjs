// Display projection for the Markdown renderer. Validators use openrpc.json.
const fs = require('node:fs');

// The renderer treats oneOf branches as complete schemas, ignoring siblings.
// Materialize their shared fields for display; the canonical intersection stays
// in openrpc.json. Return false for visibly contradictory discriminator branches.
function withCommon(common, branch) {
  if (branch === false || common === false) return false;
  if (branch === true) return common;
  if (common === true) return branch;
  if (common.type && branch.type && common.type !== branch.type) return false;
  const out = {...common, ...branch};
  if (common.enum && branch.const !== undefined && !common.enum.includes(branch.const)) return false;
  if (branch.enum && common.const !== undefined && !branch.enum.includes(common.const)) return false;
  if (common.const !== undefined && branch.const !== undefined && common.const !== branch.const) return false;
  if (common.properties || branch.properties) {
    out.properties = {...common.properties, ...branch.properties};
    for (const key of Object.keys(common.properties || {})) {
      if (branch.properties?.[key]) {
        out.properties[key] = withCommon(common.properties[key], branch.properties[key]);
        if (out.properties[key] === false && [...(common.required || []), ...(branch.required || [])].includes(key)) return false;
      }
    }
  }
  if (common.required || branch.required) out.required = [...new Set([...(common.required || []), ...(branch.required || [])])];
  const descriptions = [...new Set([common.description, branch.description].filter(Boolean))];
  if (descriptions.length) out.description = descriptions.join('\n\n');
  return out;
}

function forDisplay(schema, resource) {
  if (schema === null || typeof schema !== 'object') return schema;
  resource = schema.$id || resource;
  if (schema.$ref === '#') {
    if (!resource) throw new Error('Unscoped recursive schema reference');
    return {
      type: 'object',
      description: `Recursive instance of ${resource}. The child has the same schema as its parent resource; see the canonical OpenRPC schema for validation.`,
    };
  }
  // Once a union branch fixes a discriminator, its conditional localization
  // properties can be shown directly. Leave undecidable conditions untouched.
  if (schema.if?.properties) {
    const conditions = Object.entries(schema.if.properties);
    if (conditions.length && conditions.every(([key, value]) => value.const !== undefined && schema.properties?.[key]?.const !== undefined)) {
      const matches = conditions.every(([key, value]) => schema.properties[key].const === value.const);
      const {if: condition, then, else: otherwise, ...base} = schema;
      return forDisplay(withCommon(base, (matches ? then : otherwise) || {}), resource);
    }
  }
  const union = schema.oneOf ? 'oneOf' : schema.anyOf ? 'anyOf' : undefined;
  if (union) {
    const {[union]: variants, ...common} = schema;
    const projected = variants.map(branch => withCommon(common, branch))
      .filter(branch => branch !== false).map(branch => forDisplay(branch, resource));
    // Flatten nested unions after shared object fields have reached each leaf.
    const branches = projected.flatMap(branch => branch.oneOf || branch.anyOf || [branch]);
    if (branches.length === 1) return branches[0];
    return {title: schema.title, description: schema.description, [union]: branches};
  }
  const out = {...schema};
  // JSON Schema permits const/enum without type; the renderer does not.
  const literal = out.const ?? out.enum?.[0];
  if (!out.type && literal !== undefined && (out.const !== undefined || out.enum.every(value => typeof value === typeof literal))) {
    out.type = literal === null ? 'null' : typeof literal;
  }
  if (out.not?.required?.length === 1 && Object.keys(out.not).length === 1) {
    const forbidden = out.not.required[0];
    out.properties = {...out.properties};
    delete out.properties[forbidden];
  }
  for (const key of ['properties', 'patternProperties', 'definitions', '$defs']) {
    if (out[key]) out[key] = Object.fromEntries(Object.entries(out[key]).map(([name, child]) => [name, forDisplay(child, resource)]));
  }
  for (const key of ['items', 'additionalProperties', 'not', 'if', 'then', 'else']) {
    if (out[key] && typeof out[key] === 'object') {
      out[key] = Array.isArray(out[key]) ? out[key].map(child => forDisplay(child, resource)) : forDisplay(out[key], resource);
    }
  }
  // The renderer drops the array label when items is a union. Display each
  // element variant in an array wrapper, explicitly allowing mixed variants.
  // This is a presentation device, not an equivalent validation schema.
  const itemUnion = out.items?.oneOf ? 'oneOf' : out.items?.anyOf ? 'anyOf' : undefined;
  if (out.type === 'array' && itemUnion) {
    const description = [out.description, 'Array result. Each item independently matches one of the following variants; different variants may occur in the same array.'].filter(Boolean).join('\n\n');
    return {title: out.title, description, [itemUnion]: out.items[itemUnion].map(items => ({...out, title: items.title || out.title, description, items}))};
  }
  return out;
}

// Keep existing method documentation stable while these composition workarounds
// are exercised by the new trace schemas. Recursion remains bounded everywhere.
function abbreviateRecursion(value, resource) {
  if (Array.isArray(value)) return value.map(child => abbreviateRecursion(child, resource));
  if (value === null || typeof value !== 'object') return value;
  resource = value.$id || resource;
  if (value.$ref === '#') return forDisplay(value, resource);
  return Object.fromEntries(Object.entries(value).map(([key, child]) => [key, abbreviateRecursion(child, resource)]));
}

function documentForDisplay(document) {
  // Examples are data, not schemas: do not rewrite objects that happen to have
  // schema keyword names in an example value.
  return {
    ...document,
    methods: document.methods.map(method => {
      const project = method.name?.startsWith('trace_') ? forDisplay : abbreviateRecursion;
      return {
        ...method,
        params: method.params.map(param => ({...param, schema: project(param.schema)})),
        result: {...method.result, schema: project(method.result.schema)},
        ...(method.errors && {
          errors: method.errors.map(error => error.data ? {...error, data: project(error.data)} : error),
        }),
      };
    }),
  };
}

if (require.main === module) {
  const document = JSON.parse(fs.readFileSync('openrpc.json', 'utf8'));
  fs.writeFileSync('docs-openrpc.json', JSON.stringify(documentForDisplay(document), null, 2) + '\n');
}
module.exports = {forDisplay, documentForDisplay};
