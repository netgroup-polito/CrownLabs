const { createGraphQlSchema } = require('oasgraph');
const { decorateBaseSchema } = require('./decorateBaseSchema.js');

exports.createSchema = async (oas, kubeApiUrl, token) => {
  let baseSchema = await oasToGraphQlSchema(oas, kubeApiUrl, token);
  try {
    const schemaWithInstanceTemplate = decorateBaseSchema(
      'itPolitoCrownlabsV1alpha2Template',
      'TemplateCrownlabsPolitoItTemplateRef',
      baseSchema,
      'templateWrapper',
      ['name', 'namespace']
    );
    const schemaWithInstanceTenant = decorateBaseSchema(
      'itPolitoCrownlabsV1alpha1Tenant',
      'TenantCrownlabsPolitoItTenantRef',
      schemaWithInstanceTemplate,
      'tenantWrapper',
      ['name']
    );

    return schemaWithInstanceTenant;
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
      'Content-Type': 'application/json',
    },
    tokenJSONpath: '$.token',
  });
  return schema;
}
