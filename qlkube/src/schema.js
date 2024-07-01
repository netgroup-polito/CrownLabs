const { createGraphQLSchema } = require('openapi-to-graphql');
const { stitchSchemas } = require('@graphql-tools/stitch');
const { RenameRootFields } = require('@graphql-tools/wrap');
const { RenameRootTypes } = require('@graphql-tools/wrap');
const { decorateBaseSchema } = require('./decorateBaseSchema');
const { wrappers } = require('./wrappers');

const basicHeaders = {
  'Content-Type': 'application/json',
};

async function oasToGraphQlSchema(oas, baseUrl, token, operationIdFieldNames) {
  const schema = await createGraphQLSchema(oas, {
    baseUrl,
    viewer: false,
    headers: token ? {
      Authorization: `Bearer ${token}`,
      ...basicHeaders,
    } : basicHeaders,
    tokenJSONpath: '$.token',
    simpleEnumValues: true,
    operationIdFieldNames,
  });
  return schema;
}

exports.createSchema = async (oas, kubeApiUrl, token) => {
  let baseSchema = (await oasToGraphQlSchema(oas, kubeApiUrl, token)).schema;

  wrappers.forEach(
    ({
      type, fieldWrapper, nameWrapper, queryFieldsRequired,
    }) => {
      baseSchema = decorateBaseSchema(
        type,
        fieldWrapper,
        baseSchema,
        nameWrapper,
        queryFieldsRequired,
      );
    },
  );

  return baseSchema;
};

exports.oasToGraphQlSchema = oasToGraphQlSchema;

/**
 * stitch schemas by merging them, possibly prefixing root fields
 * schemas is an object like { schema, prefix? }
 *
 * @param {GraphQLSchema[]} schemas
 * @returns
 */
module.exports.joinSchemas = (schemas) => {
  const subschemas = schemas.map(({ schema, prefix }) => ({
    schema,
    transforms: prefix ? [
      new RenameRootFields((operation, name, _field) => `${prefix}_${name}`),
      new RenameRootTypes((name) => `${prefix}_${name}`),
    ] : [],
  }));

  const schema = stitchSchemas({ subschemas });
  return schema;
};
