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
        }
        status {
          ip
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
          templateCrownlabsPolitoItTemplateRef {
            name
            namespace
            templateWrapper {
              itPolitoCrownlabsV1alpha2Template {
                spec {
                  prettyName
                  description
                  allowPublicExposure
                  environmentList {
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
