import type { FC } from 'react';
import { Table, Tag } from 'antd';
import type { TenantQuery } from '../../../generated-types';

export interface ITenantWorkspacesProps {
  tenant: TenantQuery;
}

const TenantWorkspaces: FC<ITenantWorkspacesProps> = ({ tenant }) => {
  const workspaces = (tenant.tenant?.spec?.workspaces ?? [])
    .map(w => ({
      name:
        w?.workspaceWrapperTenantV1alpha2?.itPolitoCrownlabsV1alpha1Workspace
          ?.spec?.prettyName ??
        w?.name ??
        '',
      role: w?.role ?? 'unknown',
    }))
    .sort((a, b) => a.name.localeCompare(b.name));

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
