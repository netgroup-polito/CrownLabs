import gql from 'graphql-tag';

export default gql`
  subscription updatedSshKeys($tenantId: String!) {
    updatedTenant: itPolitoCrownlabsV1alpha1TenantUpdate(name: $tenantId) {
      updatedKeys: payload {
        spec {
          email
          firstName
          lastName
          publicKeys
        }
      }
    }
  }
`;
