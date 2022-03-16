import fs from 'fs';
import { buildClientSchema, printSchema, getIntrospectionQuery } from 'graphql';
import { request } from 'graphql-request';

const endpoint = 'http://localhost:8545/graphql';

async function main() {
    const q = getIntrospectionQuery();
    const res = await request(endpoint, q);
    const schema = JSON.stringify(res, null, 2);
    fs.writeFileSync('graphql.json', schema);
}

main().then().catch((err) => console.log(err))
