// This file serves development/testing purposes only.
// When qlkube is deployed with Helm, it is overwritten by an equivalent
// one automatically generated from the configuration therein specified.

const apiGroups = {
  paths: ['crownlabs.polito.it'],
};

const regexes = [
  /crownlabs\.polito\.it.+/,
];

const subscriptions = [
  {
    api: 'apis',
    group: 'crownlabs.polito.it',
    version: 'v1alpha2',
    resource: 'instances',
    type: 'itPolitoCrownlabsV1alpha2Instance',
    listMapping: 'itPolitoCrownlabsV1alpha2InstanceList',
  },
  {
    api: 'apis',
    group: 'crownlabs.polito.it',
    version: 'v1alpha2',
    resource: 'templates',
    type: 'itPolitoCrownlabsV1alpha2Template',
    listMapping: null,
  },
];

const wrappers = [
  {
    type: 'itPolitoCrownlabsV1alpha2Template',
    fieldWrapper: 'TemplateCrownlabsPolitoItTemplateRef',
    nameWrapper: 'templateWrapper',
    queryFieldsRequired: ['name', 'namespace'],
    parents: ['itPolitoCrownlabsV1alpha2Instance'],
  },
];

module.exports = { apiGroups, subscriptions, wrappers, regexes };
