// This file serves development/testing purposes only.
// When qlkube is deployed with Helm, it is overwritten by an equivalent
// one automatically generated from the configuration therein specified.

const wrappersRemote = [
  {
    type: 'itPolitoCrownlabsV1alpha2Template',
    fieldWrapper: 'TemplateCrownlabsPolitoItTemplateRef',
    nameWrapper: 'templateWrapper',
    queryFieldsRequired: ['name', 'namespace'],
    parents: ['itPolitoCrownlabsV1alpha2Instance'],
  },
];

const wrappersLocal = [
  {
    type: 'itPolitoCrownlabsV1alpha2Template',
    fieldWrapper: 'TemplateCrownlabsPolitoItTemplateRef',
    nameWrapper: 'templateWrapper',
    queryFieldsRequired: ['name', 'namespace'],
    parents: ['itPolitoCrownlabsV1alpha2Instance'],
  },
  {
    type: 'itPolitoCrownlabsV1alpha2Tenant',
    fieldWrapper: 'TenantCrownlabsPolitoItTenantRef',
    nameWrapper: 'tenantV1alpha2Wrapper',
    queryFieldsRequired: ['name'],
    parents: ['itPolitoCrownlabsV1alpha2Instance'],
  },
  {
    type: 'itPolitoCrownlabsV1alpha1Workspace',
    fieldWrapper: 'WorkspacesListItem',
    nameWrapper: 'workspaceWrapperTenantV1alpha2',
    queryFieldsRequired: ['name'],
    parents: ['itPolitoCrownlabsV1alpha2Tenant'],
  },
];

module.exports = { wrappers: process.env.USE_LOCAL_CLUSTER ? wrappersLocal : wrappersRemote };
