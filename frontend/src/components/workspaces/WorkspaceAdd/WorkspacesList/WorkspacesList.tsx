import type { FC } from 'react';
import type { WorkspacesAvailable } from '../../../../utils';
import { Empty, Table } from 'antd';
import { WorkspaceRow } from '../WorkspaceRow';

export interface IWorkspaceListProps {
  workspacesAvailable: WorkspacesAvailable[];
  action: (w: WorkspacesAvailable) => void;
}

const WorkspacesList: FC<IWorkspaceListProps> = ({ ...args }) => {
  const { workspacesAvailable, action } = args;

  const columns = [
    {
      title: 'Workspace',
      key: 'workspace',
      render: (record: WorkspacesAvailable) => (
        <WorkspaceRow workspace={record} action={action}></WorkspaceRow>
      ),
    },
  ];

  return workspacesAvailable.length > 0 ? (
    <div className="w-full flex-grow flex-wrap content-between py-0 overflow-auto scrollbar cl-templates-table">
      <Table
        size="middle"
        showHeader={false}
        rowKey={w => w.name}
        columns={columns}
        dataSource={workspacesAvailable}
        pagination={false}
      />
    </div>
  ) : (
    <div className="w-full h-full flex-grow flex flex-wrap content-center justify-center py-5 ">
      <div className="w-full pb-10 flex justify-center">
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={false} />
      </div>
      <p className="text-md xs:text-xl text-center px-5 xs:px-24">
        No workspaces available
      </p>
    </div>
  );
};

export default WorkspacesList;
