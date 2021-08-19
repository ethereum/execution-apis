import fs from "fs";

console.log("Loading methods...\n");

let methods = [];
let methodsBase = "src/methods/";
let methodFiles = fs.readdirSync(methodsBase);
for (const file of methodFiles) {
  let raw = fs.readFileSync(methodsBase + file);
  let parsed = JSON.parse(raw);
  methods.push(parsed);
}

console.log("Loading schemas...\n");

let schemas = {};
let schemasBase = "src/schemas/";
let schemaFiles = fs.readdirSync(schemasBase);
for(const file of schemaFiles) {
  let raw = fs.readFileSync(schemasBase + file);
  let parsed = JSON.parse(raw);
  schemas = {
    ...schemas,
    ...parsed,
  };
};

console.log("Loading descriptions...\n");

let descriptionBase = "src/description/";
let descriptionFiles = fs.readdirSync(descriptionBase);
for (const file of descriptionFiles) {
  const raw = fs.readFileSync(descriptionBase + file);
	// captures the file name before the file type
  const methodName = file.split(".")[0];
  const stringyDescription = raw.toString();
  methods = methods.map((method, i) => {
    if (method.name.toLowerCase() === methodName.toLowerCase()) {
      method.description = stringyDescription;
    }
		return method
  });
}

const spec = {
  openrpc: "1.2.4",
  info: {
    title: "Ethereum JSON-RPC Specification",
    description:
      "A specification of the standard interface for Ethereum clients.",
    license: {
      name: "CC0-1.0",
      url: "https://creativecommons.org/publicdomain/zero/1.0/legalcode",
    },
    version: "0.0.0",
  },
  methods: methods,
  components: {
    schemas: schemas,
  },
};

let data = JSON.stringify(spec, null, "\t");
fs.writeFileSync("openrpc.json", data);

console.log();
console.log("Build successful.");
