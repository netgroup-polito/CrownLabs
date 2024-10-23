const { stitchSchemas } = require('@graphql-tools/stitch');
const { RenameRootFields } = require('@graphql-tools/wrap');

module.exports.joinSchemas = (kubeSchema, harborSchema) => {
  let schema = stitchSchemas({
    subschemas: [
      {
        schema: kubeSchema,
      },
      {
        schema: harborSchema,
        transforms: [
          new RenameRootFields((operation, name, field) => {
            return 'reg_' + name;
          }),
        ],
      },
    ],
  });
  return schema;
};
