const { apiGroups } = require('./apiGroups');

const k8s = 'io.k8s.apimachinery';

const force = {
  uniqueItems: true,
  type: 'boolean',
  description:
    'Force is going to "force" Apply requests. It means user will re-acquire conflicting fields owned by other people. Force flag must be unset for non-apply patch requests.',
  name: 'force',
  in: 'query',
};

/**
 * Recursively add x-graphql-enum-mapping property
 * to the enum ones which values contain empty string
 *
 * @param {Object} obj
 */
function setGQLEnumNodes(obj) {
  const GQL_EXT_ENUM_MAPPING = 'x-graphql-enum-mapping';
  if (!obj.properties) return;
  Object.keys(obj.properties).forEach((prop) => {
    const p = obj.properties[prop];
    if (p?.type === 'string' && p.enum) {
      if (!p[GQL_EXT_ENUM_MAPPING]) {
        p[GQL_EXT_ENUM_MAPPING] = {};
      }
      p.enum.forEach((v) => {
        if (!p[GQL_EXT_ENUM_MAPPING][v]) {
          if (v === '') {
            p[GQL_EXT_ENUM_MAPPING][''] = '_EMPTY_';
          }
        }
      });
    } else if (p?.type === 'object') {
      setGQLEnumNodes(p);
    } else if (p?.type === 'array') {
      setGQLEnumNodes(p.items);
    }
  });
}

/**
 * Filters the OpenApi made from k8s API.
 * Starting from a list of paths,
 * for each path into oas checks if at least one of our paths matches with it,
 * so that paths that are never matched are removed.
 * @param {*} oas
 * @returns
 */

function decorateOpenapi(oas) {
  let result = true;
  let keys = Object.keys(oas.paths);
  let removeKeys = keys.filter((key) => {
    result = true;
    const { paths } = apiGroups;
    paths.forEach((path) => {
      result = result && !key.includes(path);
    });
    return result;
  });
  removeKeys.forEach((rk) => {
    // eslint-disable-next-line no-param-reassign
    delete oas.paths[rk];
  });

  keys = Object.keys(oas.definitions);
  removeKeys = keys.filter((key) => {
    result = true;
    const definitions = apiGroups.paths.map((definition) => definition.split('.').reverse().join('.'));
    definitions.push(k8s);
    definitions.forEach((definition) => {
      result = result && !key.includes(definition);
    });
    return result;
  });
  removeKeys.forEach((rk) => {
    // eslint-disable-next-line no-param-reassign
    delete oas.definitions[rk];
  });

  for (const path of Object.keys(oas.paths)) {
    const { patch } = oas.paths[path];
    if (patch) {
      patch.consumes = ['application/apply-patch+yaml'];
      if (!patch.parameters.find((p) => p.name === 'force')) patch.parameters.push({ ...force });
    }
  }

  Object.keys(oas.definitions).forEach((k) => setGQLEnumNodes(oas.definitions[k]));

  return oas;
}

module.exports = { decorateOpenapi };
