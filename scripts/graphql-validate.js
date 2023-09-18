import fs from 'fs';
import graphql from 'graphql';
import diff frm '@graphql-inspector/core';

function ignoreDirectiveChanges(obj) {
  return obj.changes.filter((change) => !change.type.startsWith('DIRECTIVE'));
}

// Validate JSON schema
const raw = fs.readFileSync('graphql.json');
const schema = graphql.buildClientSchema(JSON.parse(raw));
graphql.assertValidSchema(schema);
console.log('GraphQL JSON schema validated successfully.');

// Validate standard schema
const rawStd = fs.readFileSync('schema.graphqls', 'utf8');
const schemaStd = graphql.buildSchema(rawStd);
graphql.assertValidSchema(schemaStd);
console.log('GraphQL standard schema validated successfully.');

// Compare and make sure JSON and standard schemas match.
diff(schema, schemaStd, [ignoreDirectiveChanges])
  .then((changes) => {
    if (changes.length === 0) {
      console.log('GraphQL schemas match.');
      return;
    }
    throw new Error(
      `Found differences between JSON and standard:\n${JSON.stringify(
        changes,
        null,
        2
      )}`
    );
  })
  .catch(console.error);

fs.readdir('graphql/tests', (_, files) => {
  files.forEach((file) => {
    const query = graphql.parse(
      fs.readFileSync(`graphql/tests/${file}/request.gql`, 'utf8')
    );
    const output = JSON.parse(
      fs.readFileSync(`graphql/tests/${file}/response.json`, 'utf8')
    );
    if (!('statusCode' in output) || !('responses' in output)) {
      throw new Error(
        `GraphQL response ${file} without 'statusCode' or 'responses' keys`
      );
    }
    if (output['statusCode'] === 200) {
      const result = graphql.validate(schema, query);
      if (result.length === 0) {
        console.log(`GraphQL request ${file} validated successfully.`);
      } else {
        throw new Error(
          `GraphQL query ${file} failed validation:\n${JSON.stringify(
            result,
            null,
            2
          )}`
        );
      }
    }
  });
});
