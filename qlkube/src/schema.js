const { createGraphQLSchema } = require('openapi-to-graphql');
const { decorateBaseSchema } = require('./decorateBaseSchema');
const { wrappers } = require('./wrappers');

exports.createHarborSchema = async (oas, harborApiUrl, token) => {
  const { schema } = await createGraphQLSchema(oas, {
    baseUrl: `${harborApiUrl}/api/v2.0`,
    viewer: false,
    requestOptions: {
      headers: (_method, _path, _title, _resolverParams) => ({
        Authorization: `Basic ${btoa(`${token.name}:${token.secret}`)}`,
      }),
    },
  });
  return schema;
};

async function oasToGraphQlSchema(oas, kubeApiUrl, token) {
  const schema = await createGraphQLSchema(oas, {
    baseUrl: kubeApiUrl,
    viewer: false,
    headers: {
      Authorization: `Bearer ${token}`,
    },
    tokenJSONpath: '$.token',
    simpleEnumValues: true,
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
