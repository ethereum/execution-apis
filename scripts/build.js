import fs from "fs";

console.log("Loading files...\n");

let methods = [];
let methodsBase = "src/methods/";
let methodFiles = fs.readdirSync(methodsBase);
methodFiles.forEach(file => {
	console.log(file);
  let raw = fs.readFileSync(methodsBase + file);
	let parsed = JSON.parse(raw);
	methods = [
		...methods,
	  ...parsed
	]
});

let schemas = {};
let schemasBase = "src/schemas/"
let schemaFiles = fs.readdirSync(schemasBase);
schemaFiles.forEach(file => {
	console.log(file);
  let raw = fs.readFileSync(schemasBase + file);
	let parsed = JSON.parse(raw);
	schemas = {
		...schemas,
		...parsed,
	};
});

let descriptionBase = "src/description/"
let descriptionFiles = fs.readdirSync(descriptionBase);
descriptionFiles.forEach(file => {
	let raw = fs.readFileSync(descriptionBase + file);
	const methodName = file.split(".")[0]
	console.log(methodName);
	let stringyDescription = raw.toString();
	// stringyDescription === the description we want for some method
	methods.forEach((method, i) => {
		if (method.name === methodName) {
			
			methods[i].description = stringyDescription;
		}
	})
});


console.log(methods[1])

const spec = {
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

let data = JSON.stringify(spec, null, '\t');
fs.writeFileSync('openrpc.json', data);

console.log();
console.log("Build successful.");
