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
          lastLogin
          quota {
            cpu
            instances
            memory
          }
          workspaces {
            role
            name
            workspaceWrapperTenantV1alpha2 {
              itPolitoCrownlabsV1alpha1Workspace {
                spec {
                  prettyName
                  quota {
                    cpu
                    instances
                    memory
                  }
                }
                status {
                  namespace {
                    name
                  }
                }
              }
            }
          }
          publicKeys
        }
        metadata {
          name
        }
        status {
          personalNamespace {
            name
            created
          }
          quota {
            cpu
            instances
            memory
          }
        }
      }
    }
  }
`;
