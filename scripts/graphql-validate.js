import fs from 'fs';
import path from 'path';
import graphql from 'graphql';
import { diff } from '@graphql-inspector/core';

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

function validateGraphql(dir) {
  fs.readdir(dir, (_, files) => {
    files.forEach((file) => {
      const filePath = path.join(dir, file);

      console.log(`Validating file: ${filePath}`);
      if (fs.statSync(filePath).isFile()) {
        const lines = fs.readFileSync(filePath, 'utf8').split('\n');

        let prev = null;
        lines.forEach((line) => {
          if (prev && prev.startsWith('>> ') && line.startsWith('<< ')) {
            const output = JSON.parse(line.substring(3));

            // Validate the success Query
            if (!'errors' in output) {
              const query = graphql.parse(line.substring(3));
              const result = graphql.validate(schema, query);
              if (result.length === 0) {
                console.log(`GraphQL test ${file} validated successfully.`);
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
          }

          prev = line;
        });
      } else {
        validateGraphql(filePath);
      }
    });
  });
}
validateGraphql('tests/graphql');
