const { apiGroups } = require('./apiGroups');

const k8s = 'io.k8s.apimachinery';

function decorateOpenapi(oas) {
  let result = true;
  let keys = Object.keys(oas.paths);
  let removeKeys = keys.filter(key => {
    result = true;
    const paths = apiGroups.paths;
    paths.forEach(path => {
      result &&= key.indexOf(path) == -1;
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
      result &&= key.indexOf(definition) == -1;
    });
    return result;
  });
  removeKeys.forEach(rk => {
    delete oas.definitions[rk];
  });
  return oas;
}

module.exports = { decorateOpenapi };
