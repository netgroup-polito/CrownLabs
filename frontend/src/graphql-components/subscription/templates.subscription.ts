import gql from 'graphql-tag';

export default gql`
  subscription updatedWorkspaceTemplates(
    $workspaceNamespace: String!
    $templateName: String!
  ) {
    updatedTemplate: itPolitoCrownlabsV1alpha2TemplateUpdate(
      namespace: $workspaceNamespace
      name: $templateName
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
