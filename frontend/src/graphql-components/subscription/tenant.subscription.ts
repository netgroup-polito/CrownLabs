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
            workspaceId: name
            workspaceWrapper {
              itPolitoCrownlabsV1alpha1Workspace {
                spec {
                  workspaceName: prettyName
                }
                status {
                  namespace {
                    workspaceNamespace: name
                  }
                }
              }
            }
          }
        }
        metadata {
          tenantId: name
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
