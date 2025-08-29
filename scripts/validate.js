import fs from "fs";
import { 
  parseOpenRPCDocument,
  dereferenceDocument,
  validateOpenRPCDocument
} from "@open-rpc/schema-utils-js";
import OpenrpcDocument from "@open-rpc/meta-schema";

let rawdata = fs.readFileSync("openrpc.json");
let openrpc = JSON.parse(rawdata);

/** @type {OpenrpcDocument} */
const document = openrpc;
const dereffed = await dereferenceDocument(document);

const error = validateOpenRPCDocument(dereffed);
if (error != true) {
  console.log(error.name);
  console.log(error.message);
  process.exit(1);
}

try {
  await Promise.resolve(parseOpenRPCDocument(openrpc));
} catch(e) {
  console.log(e.name);
  let end = e.message.indexOf("schema in question");
  let msg = e.message.substring(0, end);
  console.log(msg);
  process.exit(1);
}

console.log("OpenRPC spec validated successfully.");
