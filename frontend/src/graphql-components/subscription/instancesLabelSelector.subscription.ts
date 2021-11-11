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
        }
        spec {
          running
          prettyName
          tenantCrownlabsPolitoItTenantRef {
            tenantId: name
            tenantWrapper {
              itPolitoCrownlabsV1alpha1Tenant {
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
                  templateName: prettyName
                  templateDescription: description
                  environmentList {
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
