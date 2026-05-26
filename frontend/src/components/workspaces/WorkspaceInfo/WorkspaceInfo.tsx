import type { FC } from 'react';
import type { WorkspaceQuery } from '../../../generated-types';

export interface IWorkspaceInfoProps {
  workspace: WorkspaceQuery;
}

const WorkspaceInfo: FC<IWorkspaceInfoProps> = ({ workspace }) => {
  return (
    <>
      <h3>Workspace Details</h3>
      <p>
        Name: <strong>{workspace.workspace?.metadata?.name}</strong>
      </p>
      <p>
        Pretty Name: <strong>{workspace.workspace?.spec?.prettyName}</strong>
      </p>
      <p>
        Auto Enroll: <strong>{workspace.workspace?.spec?.autoEnroll}</strong>
      </p>
      <p>
        Namespace: <strong>{workspace.workspace?.metadata?.namespace}</strong>
      </p>
    </>
  );
};

export default WorkspaceInfo;
