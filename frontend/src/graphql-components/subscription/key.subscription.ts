import gql from 'graphql-tag';

export default gql`
  subscription updatedSshKeys($tenantId: String!) {
    updatedTenant: itPolitoCrownlabsV1alpha2TenantUpdate(name: $tenantId) {
      updateType
      updatedKeys: payload {
        metadata {
          name
        }
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
