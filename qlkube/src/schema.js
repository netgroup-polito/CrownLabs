const { createGraphQLSchema } = require('openapi-to-graphql');
const { decorateBaseSchema } = require('./decorateBaseSchema.js');
const { wrappers } = require('./wrappers');

exports.createSchema = async (oas, kubeApiUrl, token) => {
  let baseSchema = (await oasToGraphQlSchema(oas, kubeApiUrl, token)).schema;

  try {
    wrappers.forEach(
      ({ type, fieldWrapper, nameWrapper, queryFieldsRequired }) => {
        baseSchema = decorateBaseSchema(
          type,
          fieldWrapper,
          baseSchema,
          nameWrapper,
          queryFieldsRequired
        );
      }
    );

    return baseSchema;
  } catch (e) {
    console.error(e);
    process.exit(1);
  }
};

exports.createHarborSchema = async (oas, harborApiUrl, token) => {
  let { schema } = await createGraphQLSchema(oas, {
    baseUrl: harborApiUrl + '/api/v2.0',
    viewer: false,
    requestOptions: {
      headers: (method, path, title, resolverParams) => {
        console.log(method, path, title);
        return {
          Authorization: 'Basic ' + btoa(token.name + ':' + token.secret),
        };
      },
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
