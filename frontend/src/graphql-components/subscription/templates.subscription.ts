import gql from 'graphql-tag';
import { TemplateEnvironmentFieldsFragmentDoc } from '../../generated-types';

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
          allowPublicExposure
          nodeSelector
          cleanup {
            deleteAfterCreation
            stopAfterInactivity
            deleteAfterInactivity
          }
          environmentList {
            ...TemplateEnvironmentFields
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
  ${TemplateEnvironmentFieldsFragmentDoc}
`;
