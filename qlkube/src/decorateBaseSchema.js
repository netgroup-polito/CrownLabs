const { gql } = require('apollo-server-core');
const { extendSchema } = require('graphql/utilities');
const { addResolversToSchema } = require('@graphql-tools/schema');
const { capitalizeType, graphqlLogger } = require('./utils.js');

/**
 *
 * @param {*} targetQuery : Query wrapped int the extended object Type
 * @param {*} extendedType : Object type that should be extended
 * @param {*} baseSchema : GraphQL Schema where the type and query are written
 * @param {*} nameWrapper : Custom name for the wrapper
 * @param {*} argsNeeded : Arguments extracted from the father and passed from the wrapper
 * @returns
 */

function decorateBaseSchema(
  targetQuery,
  extendedType,
  baseSchema,
  nameWrapper,
  argsNeeded
) {
  if (!targetQuery) throw new Error('Parameter targetQuery cannot be empty!');
  if (!extendedType) throw new Error('Parameter extendedType cannot be empty!');
  if (!nameWrapper) throw new Error('Parameter nameWrapper cannot be empty!');
  if (!argsNeeded) throw new Error('Parameter argsNeeded cannot be empty!');
  if (!baseSchema) throw new Error('Parameter baseSchema cannot be empty!');

  if (baseSchema.getQueryType().getFields()[targetQuery] === undefined)
    throw new Error('Parameter targetQuery not valid!');
  const targetType = baseSchema.getQueryType().getFields()[targetQuery];

  if (!targetType) throw new Error('targetType fault!');

  const typeWrapper = capitalizeType(nameWrapper);
  const typeTargetQuery = capitalizeType(targetQuery);

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
        const newParent = {};
        for (e of argsNeeded) {
          if (parent[e] === undefined)
            throw new Error(`Error: ${e} is not into the parent object!`);
          newParent[e] = parent[e];
        }
        return newParent;
      },
    },
    [typeWrapper]: {
      [targetQuery]: (parent, args, context, info) => {
        graphqlLogger(
          `[i] Resolve ${targetQuery} query of ${typeWrapper} type wrapper`
        );
        return targetType.resolve(parent, parent, context, info);
      },
    },
  };

  const extendedSchema = extendSchema(baseSchema, extension);
  const newSchema = addResolversToSchema(extendedSchema, resolvers);
  return newSchema;
}

module.exports = { decorateBaseSchema };
