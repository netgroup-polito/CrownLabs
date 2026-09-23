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
          personalWorkspace {
            cpu
            instances
            memory
            disk
            otherResources
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
                    disk
                    otherResources
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
          creationTimestamp
          labels
        }
        status {
          personalNamespace {
            name
            created
          }
        }
      }
    }
  }
`;
