import fs from 'fs';
import graphql from 'graphql';
import { diff, DiffRule } from '@graphql-inspector/core';

function ignoreDirectiveChanges(obj) {
    return obj.changes.filter((change) => !change.type.startsWith('DIRECTIVE'))
}

// Validate JSON schema
const raw = fs.readFileSync('graphql.json');
const schema = graphql.buildClientSchema(JSON.parse(raw));
graphql.assertValidSchema(schema)
console.log('GraphQL JSON schema validated successfully.')

// Validate standard schema
const rawStd = fs.readFileSync('schema.graphqls', 'utf8');
const schemaStd = graphql.buildSchema(rawStd);
graphql.assertValidSchema(schemaStd);
console.log('GraphQL standard schema validated successfully.');

// Compare and make sure JSON and standard schemas match.
diff(schema, schemaStd, [ignoreDirectiveChanges]).then((changes) => {
    if (changes.length === 0) {
        console.log('GraphQL schemas match.')
        return
    }
    throw new Error(`Found differences between JSON and standard:\n${JSON.stringify(changes, null, 2)}`)
}).catch(console.error);