import type { FC } from 'react';
import { Table, Tag } from 'antd';
import type { TenantQuery } from '../../../generated-types';

export interface ITenantWorkspacesProps {
  tenant: TenantQuery;
}

const TenantWorkspaces: FC<ITenantWorkspacesProps> = ({ tenant }) => {
  const workspaces = tenant.tenant?.spec?.workspaces ?? [];

  const columns = [
    {
      title: 'Workspace',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'Role',
      dataIndex: 'role',
      key: 'role',
      render: (role: string) => (
        <Tag color={role === 'manager' ? 'blue' : 'default'}>{role}</Tag>
      ),
    },
  ];

  return (
    <Table
      dataSource={workspaces}
      columns={columns}
      rowKey="name"
      pagination={false}
    />
  );
};

export default TenantWorkspaces;
