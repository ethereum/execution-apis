import fs from "fs";
import { validateOpenRPCDocument } from "@open-rpc/schema-utils-js";

let rawdata = fs.readFileSync("openrpc.json");
let openrpc = JSON.parse(rawdata);

const error = validateOpenRPCDocument(openrpc);
if (error != true) {
	console.log(error.name);
	console.log(error.message);
	process.exit(1);
}

console.log("OpenRPC spec validated successfully.");
