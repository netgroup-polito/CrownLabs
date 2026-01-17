import { Row, Col, Typography, Space } from 'antd';
import {
  DesktopOutlined,
  CloudOutlined,
  DatabaseOutlined,
} from '@ant-design/icons';
import type { FC } from 'react';
import { useContext } from 'react';
import './QuotaDisplay.less';
import {
  OwnedInstancesContext,
  type IQuota,
} from '../../../contexts/OwnedInstancesContext';

const { Text } = Typography;

export interface IQuotaDisplayProps {
  workspaceName?: string;
}

const QuotaDisplay: FC<IQuotaDisplayProps> = ({ workspaceName }) => {
  const { consumedQuota, totalQuota } = useContext(OwnedInstancesContext);

  const workspaceConsumedQuota: IQuota = consumedQuota[workspaceName || ''] || {
    instances: 0,
    cpu: 0,
    memory: 0,
    disk: 0,
  };
  const workspaceTotalQuota: IQuota = totalQuota[workspaceName || ''] || {
    instances: 0,
    cpu: 0,
    memory: 0,
    disk: 0,
  };

  return (
    <div className="quota-display-container h-25 md:h-10 px-5 h-full overflow-hidden">
      <Row gutter={[16, 0]} style={{ height: '100%' }}>
        <Col xs={24} md={8} className="md:h-full">
          <div
            className="quota-metric"
            style={{ height: '100%', display: 'flex', alignItems: 'center' }}
          >
            <Space size="small">
              <DesktopOutlined className="primary-color-fg" />
              <Text strong>
                {workspaceConsumedQuota.cpu}/{workspaceTotalQuota.cpu}
              </Text>
              <Text type="secondary" style={{ fontSize: '12px' }}>
                CPU cores
              </Text>
            </Space>
          </div>
        </Col>

        <Col xs={24} md={8} className="md:h-full">
          <div
            className="quota-metric"
            style={{ height: '100%', display: 'flex', alignItems: 'center' }}
          >
            <Space size="small">
              <DatabaseOutlined className="success-color-fg" />
              <Text strong>
                {workspaceConsumedQuota.memory.toFixed(1)}/
                {workspaceTotalQuota.memory.toFixed(1)}
              </Text>
              <Text type="secondary" style={{ fontSize: '12px' }}>
                RAM GB
              </Text>
            </Space>
          </div>
        </Col>

        <Col xs={24} md={8} className="md:h-full">
          <div
            className="quota-metric"
            style={{ height: '100%', display: 'flex', alignItems: 'center' }}
          >
            <Space size="small">
              <CloudOutlined className="warning-color-fg" />
              <Text strong>
                {workspaceConsumedQuota.instances}/
                {workspaceTotalQuota.instances}
              </Text>
              <Text type="secondary" style={{ fontSize: '12px' }}>
                Instances
              </Text>
            </Space>
          </div>
        </Col>
      </Row>
    </div>
  );
};

export default QuotaDisplay;
