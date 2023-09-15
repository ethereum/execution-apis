import fs from 'fs';
import graphql from 'graphql';

// Validate JSON schema
const raw = fs.readFileSync('graphql/schemas/schema.gql', 'utf8');
const schema = graphql.buildSchema(raw);
graphql.assertValidSchema(schema);
console.log('GraphQL schema validated successfully.');

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
