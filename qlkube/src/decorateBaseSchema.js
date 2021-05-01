const { gql } = require('apollo-server-core');
const { extendSchema } = require('graphql/utilities');
const { addResolversToSchema } = require('@graphql-tools/schema');

/**
 *
 * @param {*} targetQuery : Query wrapped int the extended object Type
 * @param {*} extendedType : Object type that should be extended
 * @param {*} argsNeeded : Optional arguments extracted from the father and passed from the wrapper
 * @param {*} baseSchema : GraphQL Schema where the type and query are written
 * @param {*} nameWrapper : Optional custom name for the wrapper
 * @returns
 */

function capitalizeType(name) {
  return name[0].toUpperCase() + name.slice(1);
}

module.exports = {
  decorateBaseSchema: function (
    targetQuery,
    extendedType,
    argsNeeded,
    nameWrapper,
    baseSchema
  ) {
    if (!targetQuery) return baseSchema;
    if (!extendedType) return baseSchema;
    const targetType = baseSchema.getQueryType().getFields()[targetQuery];
    if (!targetType) return baseSchema;

    nameWrapper = nameWrapper ? nameWrapper : 'fieldWrapper';
    let typeWrapper = capitalizeType(nameWrapper);
    let typeTargetQuery = capitalizeType(targetQuery);
    const extension = gql`
            extend type ${extendedType} {
                ${nameWrapper}: ${typeWrapper}
            }
            type ${typeWrapper} {
                ${targetQuery}: ${typeTargetQuery}
            }
        `;
    const resolvers = {
      [extendedType]: {
        [nameWrapper]: (parent, args, context, info) => {
          let newParent = {};
          for (e of argsNeeded) {
            newParent[e] = parent[e];
          }
          return newParent !== {} ? newParent : parent; // gestire errori
        },
      },
      [typeWrapper]: {
        [targetQuery]: (parent, args, context, info) => {
          return targetType.resolve(parent, parent, context, info);
        },
      },
    };

    // extending the schema
    const extendedSchema = extendSchema(baseSchema, extension);
    const newSchema = addResolversToSchema(extendedSchema, resolvers);
    return newSchema;
  },
};
