import { useEffect } from 'react';
import { Row, Col, Avatar, Tabs, Spin, Tooltip, Button } from 'antd';
import { LeftOutlined, UserOutlined } from '@ant-design/icons';
import { generateAvatarUrl } from '../../../utils';
import { useTenantLazyQuery } from '../../../generated-types';
import TenantInfo from '../TenantInfo';
import TenantPersonalWorkspaceSettings from '../TenantPersonalWorkspaceSettings';
import { Link, useParams } from 'react-router-dom';
import Box from '../../common/Box';

export default function TenantPage() {
  const { tenantId } = useParams();

  const [loadTenant, { data, loading, error }] = useTenantLazyQuery();
  useEffect(() => {
    if (tenantId) loadTenant({ variables: { tenantId } });
  }, [loadTenant, tenantId]);

  return (
    <Col span={24} lg={22} xxl={20} className="h-full">
      <Box
        header={{
          size: 'large',
          left: (
            <div className="h-full flex-none flex justify-center items-center w-20">
              <Tooltip title="Back">
                <Link to="/tenants">
                  <Button
                    type="primary"
                    shape="circle"
                    size="large"
                    icon={<LeftOutlined />}
                  />
                </Link>
              </Tooltip>
            </div>
          ),
          right: (
            <div className="h-full flex-none flex justify-center items-center w-20"></div>
          ),
          center: (
            <div className="h-full flex flex-col justify-center items-center gap-4">
              <p className="md:text-2xl text-lg text-center mb-0">
                <b>Manage user {tenantId}</b>
              </p>
            </div>
          ),
        }}
      >
        <Spin spinning={loading || error != null || !data}>
          {data && (
            <Row className="h-full w-full p-4" align="top">
              <Col xs={24} sm={8} className="text-center">
                <Avatar
                  size={100}
                  src={generateAvatarUrl(
                    'bottts',
                    data?.tenant?.spec?.email ?? '',
                  )}
                  icon={<UserOutlined />}
                />
                <p>
                  {data?.tenant?.spec?.firstName} {data?.tenant?.spec?.lastName}
                  <br />
                  <strong>{data?.tenant?.spec?.email}</strong>
                </p>
              </Col>
              <Col xs={24} sm={16} className="px-4 ">
                <Tabs
                  items={[
                    {
                      key: 'info',
                      label: 'Info',
                      children: <TenantInfo tenant={data} />,
                    },
                    {
                      key: 'personal-workspace',
                      label: 'Personal Workspace',
                      children: (
                        <TenantPersonalWorkspaceSettings tenant={data} />
                      ),
                    },
                  ]}
                />
              </Col>
            </Row>
          )}
        </Spin>
      </Box>
    </Col>
  );
}
