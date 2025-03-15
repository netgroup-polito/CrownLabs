const { stitchSchemas } = require('@graphql-tools/stitch');
const { RenameRootFields } = require('@graphql-tools/wrap');

module.exports.joinSchemas = (kubeSchema, harborSchema) => {
  const schema = stitchSchemas({
    subschemas: [
      {
        schema: kubeSchema,
      },
      {
        schema: harborSchema,
        transforms: [
          new RenameRootFields((operation, name, _field) => `reg_${name}`),
        ],
      },
    ],
  });
  return schema;
};
