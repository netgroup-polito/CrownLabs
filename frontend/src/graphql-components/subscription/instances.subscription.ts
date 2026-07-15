import gql from 'graphql-tag';

export default gql`
  subscription updatedOwnedInstances(
    $tenantNamespace: String!
    $instanceId: String
  ) {
    updateInstance: itPolitoCrownlabsV1alpha2InstanceUpdate(
      namespace: $tenantNamespace
      name: $instanceId
    ) {
      updateType
      instance: payload {
        metadata {
          name
          namespace
          creationTimestamp
          labels
          annotations
        }
        status {
          phase
          url
          nodeName
          nodeSelector
          publicExposure {
            externalIP
            phase
            ports {
              name
              port
              protocol
              targetPort
            }
          }
          environments {
            name
            phase
            ip
            initialReadyTime
          }
        }
        spec {
          running
          prettyName
          publicExposure {
            ports {
              name
              port
              protocol
              targetPort
            }
          }
          tenantCrownlabsPolitoItTenantRef {
            name
            tenantV1alpha2Wrapper {
              itPolitoCrownlabsV1alpha2Tenant {
                spec {
                  firstName
                  lastName
                }
              }
            }
          }
          templateCrownlabsPolitoItTemplateRef {
            name
            namespace
            templateWrapper {
              itPolitoCrownlabsV1alpha2Template {
                spec {
                  prettyName
                  description
                  allowPublicExposure
                  cleanup {
                    deleteAfterCreation
                    stopAfterInactivity
                    deleteAfterInactivity
                  }
                  environmentList {
                    name
                    guiEnabled
                    persistent
                    environmentType
                    resources {
                      cpu
                      memory
                      disk
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
`;
