// This file serves development/testing purposes only.
// When qlkube is deployed with Helm, it is overwritten by an equivalent
// one automatically generated from the configuration therein specified.

const subscriptionsRemote = [
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

const subscriptionsLocal = [
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
    resource: 'instancesnapshots',
    type: 'itPolitoCrownlabsV1alpha2InstanceSnapshot',
  },
  {
    api: 'apis',
    group: 'crownlabs.polito.it',
    version: 'v1alpha2',
    resource: 'templates',
    type: 'itPolitoCrownlabsV1alpha2Template',
  },
  {
    api: 'apis',
    group: 'crownlabs.polito.it',
    version: 'v1alpha2',
    resource: 'tenants',
    type: 'itPolitoCrownlabsV1alpha2Tenant',
  },
  {
    api: 'apis',
    group: 'crownlabs.polito.it',
    version: 'v1alpha1',
    resource: 'workspaces',
    type: 'itPolitoCrownlabsV1alpha1Workspace',
  },
  {
    api: 'apis',
    group: 'crownlabs.polito.it',
    version: 'v1alpha1',
    resource: 'imagelists',
    type: 'itPolitoCrownlabsV1alpha1ImageList',
  },
];

module.exports = {
  subscriptions: process.env.USE_LOCAL_CLUSTER ? subscriptionsLocal : subscriptionsRemote,
};
