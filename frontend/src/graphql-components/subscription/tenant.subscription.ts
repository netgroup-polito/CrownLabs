import gql from 'graphql-tag';

export default gql`
  subscription updatedTenant($tenantId: String!) {
    updatedTenant: itPolitoCrownlabsV1alpha1TenantUpdate(namespace: $tenantId) {
      tenant: payload {
        spec {
          email
          firstName
          lastName
          workspaces {
            role
            workspaceRef {
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
