import gql from 'graphql-tag';

export default gql`
  subscription updatedWorkspaceTemplates(
    $workspaceNamespace: String!
    $templateId: String
  ) {
    updatedTemplate: itPolitoCrownlabsV1alpha2TemplateUpdate(
      namespace: $workspaceNamespace
      name: $templateId
    ) {
      updateType
      template: payload {
        spec {
          prettyName
          description
          environmentList {
            guiEnabled
            persistent
            nodeSelector
            resources {
              cpu
              disk
              memory
            }
          }
          workspaceCrownlabsPolitoItWorkspaceRef {
            name
          }
        }
        metadata {
          name
          namespace
        }
      }
    }
  }
`;
