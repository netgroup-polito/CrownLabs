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
      template: payload {
        spec {
          name: prettyName
          description
          environmentList {
            guiEnabled
            persistent
            resources {
              cpu
              disk
              memory
            }
          }
        }
        metadata {
          id: name
        }
      }
    }
  }
`;
