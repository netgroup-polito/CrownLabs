import type { FC } from 'react';
import type { WorkspacesAvailable } from '../../../../utils';
import { WorkspacesAvailableAction } from '../../../../utils';
import { Button, Space } from 'antd';

export interface IWorkspaceRowProps {
  workspace: WorkspacesAvailable;
  action: (w: WorkspacesAvailable) => void;
}

const WorkspaceRow: FC<IWorkspaceRowProps> = ({ ...args }) => {
  const { workspace, action } = args;

  return (
    <div className="w-full flex items-center justify-between py-0">
      <Space size="middle">
        <label className="ml-3">{workspace.prettyName}</label>
      </Space>
      <Space size="small">
        {workspace.action === WorkspacesAvailableAction.Join ||
        workspace.action === WorkspacesAvailableAction.AskToJoin ? (
          <Button
            type="primary"
            shape="round"
            size={'middle'}
            onClick={() => action(workspace)}
          >
            {workspace.action === WorkspacesAvailableAction.Join
              ? 'Join'
              : 'Ask to join'}
          </Button>
        ) : (
          <Button type="primary" shape="round" size={'middle'} disabled={true}>
            Waiting approval
          </Button>
        )}
      </Space>
    </div>
  );
};

export default WorkspaceRow;
