import gql from 'graphql-tag';

export default gql`
  subscription updatedTenant($tenantId: String!) {
    updatedTenant: itPolitoCrownlabsV1alpha2TenantUpdate(name: $tenantId) {
      updateType
      tenant: payload {
        spec {
          email
          firstName
          lastName
          workspaces {
            role
            name
            workspaceWrapperTenantV1alpha2 {
              itPolitoCrownlabsV1alpha1Workspace {
                spec {
                  prettyName
                }
                status {
                  namespace {
                    name
                  }
                }
              }
            }
          }
        }
        metadata {
          name
        }
        status {
          personalNamespace {
            name
          }
        }
      }
    }
  }
`;
