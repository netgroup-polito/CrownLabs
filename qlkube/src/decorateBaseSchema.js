const { gql } = require('apollo-server-core');
const { extendSchema } = require('graphql/utilities');
const { addResolversToSchema } = require('@graphql-tools/schema');

/**
 *
 * @param {*} targetQuery : Query wrapped int the extended object Type
 * @param {*} extendedType : Object type that should be extended
 * @param {*} baseSchema : GraphQL Schema where the type and query are written
 * @param {*} nameWrapper : Custom name for the wrapper
 * @param {*} argsNeeded : Arguments extracted from the father and passed from the wrapper
 * @returns
 */

function capitalizeType(name) {
  return name[0].toUpperCase() + name.slice(1);
}

module.exports = {
  decorateBaseSchema: function (
    targetQuery,
    extendedType,
    baseSchema,
    nameWrapper,
    argsNeeded
  ) {
    if (!targetQuery) throw 'Parameter targetQuery cannot be empty!';
    if (!extendedType) throw 'Parameter extendedType cannot be empty!';
    if (!nameWrapper) throw 'Parameter nameWrapper cannot be empty!';
    if (!argsNeeded) throw 'Parameter argsNeeded cannot be empty!';
    if (!baseSchema) throw 'Parameter baseSchema cannot be empty!';

    if (baseSchema.getQueryType().getFields()[targetQuery] === undefined)
      throw 'Parameter targetQuery not valid!';
    const targetType = baseSchema.getQueryType().getFields()[targetQuery];

    if (!targetType) throw 'targetType fault!';

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
            if (parent[e] === undefined)
              throw `Error: ${e} is not into the parent object!`;
            newParent[e] = parent[e];
          }
          return newParent;
        },
      },
      [typeWrapper]: {
        [targetQuery]: (parent, args, context, info) => {
          return targetType.resolve(parent, parent, context, info);
        },
      },
    };

    const extendedSchema = extendSchema(baseSchema, extension);
    const newSchema = addResolversToSchema(extendedSchema, resolvers);
    return newSchema;
  },
};
