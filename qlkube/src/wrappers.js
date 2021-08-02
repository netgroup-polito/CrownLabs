// This file serves development/testing purposes only.
// When qlkube is deployed with Helm, it is overwritten by an equivalent
// one automatically generated from the configuration therein specified.

const wrappers = [
  {
    type: 'itPolitoCrownlabsV1alpha2Template',
    fieldWrapper: 'TemplateCrownlabsPolitoItTemplateRef',
    nameWrapper: 'templateWrapper',
    queryFieldsRequired: ['name', 'namespace'],
    parents: ['itPolitoCrownlabsV1alpha2Instance'],
  },
];

module.exports = { wrappers };
