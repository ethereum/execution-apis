import fs from "fs";
import yaml from "js-yaml";
import merger from "json-schema-merge-allof";
import { dereferenceDocument } from "@open-rpc/schema-utils-js";

console.log("Loading files...\n");

let methods = [];
let methodsBase = "src/eth/";
let methodFiles = fs.readdirSync(methodsBase);
methodFiles.forEach(file => {
  console.log(file);
  let raw = fs.readFileSync(methodsBase + file);
  let parsed = yaml.load(raw);
  methods = [
    ...methods,
    ...parsed,
  ];
});

let schemas = {};
let schemasBase = "src/schemas/"
let schemaFiles = fs.readdirSync(schemasBase);
schemaFiles.forEach(file => {
  console.log(file);
  let raw = fs.readFileSync(schemasBase + file);
  let parsed = yaml.load(raw);
  schemas = {
    ...schemas,
    ...parsed,
  };
});
const doc = {
  openrpc: "1.2.4",
  info: {
    title: "Ethereum JSON-RPC Specification",
    description: "A specification of the standard interface for Ethereum clients.",
    license: {
      name: "CC0-1.0",
      url: "https://creativecommons.org/publicdomain/zero/1.0/legalcode"
    },
    version: "0.0.0"
  },
  methods: methods,
  components: {
    schemas: schemas
  }
}

fs.writeFileSync('refs-openrpc.json', JSON.stringify(doc, null, '\t'));

let spec = await dereferenceDocument(doc);

spec.components = {};

function recursiveMerge(schema) {
  schema = merger(schema);

  if("items" in schema && "oneOf" in schema.items) {
      schema.items.oneOf = recursiveMerge(schema.items.oneOf);
  }
  if("oneOf" in schema) {
    for(var k=0; k < schema.oneOf.length; k++) {
      schema.oneOf[k] = recursiveMerge(schema.oneOf[k]);
    }
  }
  return schema;
}

// Merge instances of `allOf` in methods.
for (var i=0; i < spec.methods.length; i++) {
  for (var j=0; j < spec.methods[i].params.length; j++) {
    spec.methods[i].params[j].schema = recursiveMerge(spec.methods[i].params[j].schema);
  }
  spec.methods[i].result.schema = recursiveMerge(spec.methods[i].result.schema);
}

let data = JSON.stringify(spec, null, '\t');
fs.writeFileSync('openrpc.json', data);

console.log();
console.log("Build successful.");
