import gql from 'graphql-tag';
import { WorkspaceImageFieldsFragmentDoc } from '../../generated-types';

export default gql`
  subscription updatedWorkspaceImages($workspaceNamespace: String!) {
    updatedImage: itPolitoCrownlabsV1alpha2InstanceSnapshotUpdate(
      namespace: $workspaceNamespace
    ) {
      updateType
      image: payload {
        ...WorkspaceImageFields
      }
    }
  }
  ${WorkspaceImageFieldsFragmentDoc}
`;
