import gql from 'graphql-tag';

export default gql`
  subscription updatedInstancesLabelSelector($labels: String) {
    updateInstanceLabelSelector: itPolitoCrownlabsV1alpha2InstanceLabelsUpdate(
      labelSelector: $labels
    ) {
      updateType
      instance: payload {
        metadata {
          name
          namespace
          creationTimestamp
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
                  environmentList {
                    name
                    guiEnabled
                    persistent
                    environmentType
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
