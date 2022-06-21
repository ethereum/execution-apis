import fs from 'fs';
import graphql from 'graphql';

const raw = fs.readFileSync('graphql.json');
const schema = graphql.buildClientSchema(JSON.parse(raw));
graphql.assertValidSchema(schema)
console.log('GraphQL schema validated successfully.')
