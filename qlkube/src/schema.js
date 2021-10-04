const { createGraphQlSchema } = require('oasgraph');
const { decorateBaseSchema } = require('./decorateBaseSchema.js');
const { wrappers } = require('./wrappers');

exports.createSchema = async (oas, kubeApiUrl, token) => {
  let baseSchema = await oasToGraphQlSchema(oas, kubeApiUrl, token);
  try {
    wrappers.forEach(wtype => {
      baseSchema = decorateBaseSchema(
        wtype['type'],
        wtype['fieldWrapper'],
        baseSchema,
        wtype['nameWrapper'],
        wtype['queryFieldsRequired']
      );
    });

    return baseSchema;
  } catch (e) {
    console.error(e);
    process.exit(1);
  }
};

async function oasToGraphQlSchema(oas, kubeApiUrl, token) {
  const { schema } = await createGraphQlSchema(oas, {
    baseUrl: kubeApiUrl,
    viewer: false,
    headers: {
      Authorization: `Bearer ${token}`,
    },
    tokenJSONpath: '$.token',
  });
  return schema;
}
