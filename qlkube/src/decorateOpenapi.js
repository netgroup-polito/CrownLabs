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
 *
 * @param {*} oas
 * This function filters the OpenApi made from k8s API.
 * Starting from a list of paths,
 * for each path into oas checks if at least one of our paths matches with it,
 * so that paths that are never matched are removed.
 * @returns
 */

function decorateOpenapi(oas) {
  let result = true;
  let keys = Object.keys(oas.paths);
  let removeKeys = keys.filter(key => {
    result = true;
    const paths = apiGroups.paths;
    paths.forEach(path => {
      result = result && !key.includes(path);
    });
    return result;
  });
  removeKeys.forEach(rk => {
    delete oas.paths[rk];
  });

  keys = Object.keys(oas.definitions);
  removeKeys = keys.filter(key => {
    result = true;
    let definitions = apiGroups.paths.map(definition =>
      definition.split('.').reverse().join('.')
    );
    definitions.push(k8s);
    definitions.forEach(definition => {
      result = result && !key.includes(definition);
    });
    return result;
  });
  removeKeys.forEach(rk => {
    delete oas.definitions[rk];
  });

  for (const path in oas.paths) {
    const { patch } = oas.paths[path];
    if (patch) {
      patch.consumes = ['application/apply-patch+yaml'];
      if (!patch.parameters.find(p => p.name === 'force'))
        patch.parameters.push({ ...force });
    }
  }
  return oas;
}

module.exports = { decorateOpenapi };
