import { type FC } from 'react';
import { Row, Col, Avatar, Tabs } from 'antd';
import { UserOutlined } from '@ant-design/icons';
import { generateAvatarUrl } from '../../../utils';
import type { TenantQuery } from '../../../generated-types';
import TenantInfo from '../TenantInfo';
import TenantPersonalWorkspaceSettings from '../TenantPersonalWorkspaceSettings';

export interface ITenantPanelProps {
  tenant: TenantQuery;
}

const TenantPanel: FC<ITenantPanelProps> = ({ tenant }) => {
  return (
    <Row className="h-full w-full p-4" align="top">
      <Col xs={24} sm={8} className="text-center">
        <Avatar
          size={100}
          src={generateAvatarUrl('bottts', tenant.tenant?.spec?.email ?? '')}
          icon={<UserOutlined />}
        />
        <p>
          {tenant.tenant?.spec?.firstName} {tenant.tenant?.spec?.lastName}
          <br />
          <strong>{tenant.tenant?.spec?.email}</strong>
        </p>
      </Col>
      <Col xs={24} sm={16} className="px-4 ">
        <Tabs
          items={[
            {
              key: 'info',
              label: 'Info',
              children: <TenantInfo tenant={tenant} />,
            },
            {
              key: 'personal-workspace',
              label: 'Personal Workspace',
              children: <TenantPersonalWorkspaceSettings tenant={tenant} />,
            },
          ]}
        ></Tabs>
      </Col>
    </Row>
  );
};

export default TenantPanel;
