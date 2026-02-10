import fs from "fs";
import mergeAllOf from "json-schema-merge-allof";
import { dereferenceDocument } from "@open-rpc/schema-utils-js";

console.log("Loading spec...\n");

let rawdata = fs.readFileSync("refs-openrpc.json");
let doc = JSON.parse(rawdata);

let spec = await dereferenceDocument(doc);

spec.components = {};

// Merge instances of `allOf` in methods.
for (var i = 0; i < spec.methods.length; i++) {
  for (var j = 0; j < spec.methods[i].params.length; j++) {
    spec.methods[i].params[j].schema = mergeAllOf(spec.methods[i].params[j].schema);
  }
  spec.methods[i].result.schema = mergeAllOf(spec.methods[i].result.schema);
}

let data = JSON.stringify(spec, null, '\t');
fs.writeFileSync('openrpc.json', data);

console.log();
console.log("Build successful.");
