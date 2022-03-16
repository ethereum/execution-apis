import fs from 'fs';
import graphql from 'graphql';
import { request } from 'graphql-request';

const endpoint = 'http://localhost:8545/graphql';

async function main() {
    const q = graphql.getIntrospectionQuery();
    const res = await request(endpoint, q);
    const schema = JSON.stringify(res, null, 2);
    fs.writeFileSync('graphql.json', schema);
    console.log('GraphQL schema generated successfully.')
}

main().then().catch((err) => console.log(err));
