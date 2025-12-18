import { Row, Col, Typography, Space } from 'antd';
import {
  DesktopOutlined,
  CloudOutlined,
  DatabaseOutlined,
} from '@ant-design/icons';
import type { FC } from 'react';
import { useMemo } from 'react';
import { parseMemoryToGB } from './useQuotaCalculation';
import './QuotaDisplay.less';

const { Text } = Typography;

export interface IQuotaDisplayProps {
  consumedQuota: {
    cpu?: string | number;
    memory?: string;
    instances?: number;
  } | null;
  workspaceQuota: {
    cpu?: string | number;
    memory?: string;
    instances?: number;
  };
}

const QuotaDisplay: FC<IQuotaDisplayProps> = ({
  consumedQuota,
  workspaceQuota,
}) => {
  const quota = workspaceQuota;

  const currentUsage = useMemo(() => {
    let usedCpu = 0;
    let usedMemory = 0;
    let runningInstances = 0;

    if (consumedQuota) {
      usedCpu =
        typeof consumedQuota.cpu === 'string'
          ? parseFloat(consumedQuota.cpu) || 0
          : consumedQuota.cpu || 0;
      usedMemory = parseMemoryToGB(consumedQuota.memory || '0');
      runningInstances = consumedQuota.instances || 0;
    }

    return {
      cpu: usedCpu,
      memory: usedMemory,
      instances: runningInstances,
    };
  }, [consumedQuota]);

  const quotaLimits = {
    cpu: quota?.cpu ? parseInt(String(quota.cpu)) : 8,
    memory: quota?.memory ? parseMemoryToGB(quota.memory) : 16,
    instances: quota?.instances || 8,
  };

  return (
    <div
      className="quota-display-container h-25 md:h-10 px-5 h-full overflow-hidden"
    >
      <Row gutter={[16, 0]} style={{ height: '100%' }}>
        <Col xs={24} md={8} className='md:h-full'>
          <div
            className="quota-metric"
            style={{ height: '100%', display: 'flex', alignItems: 'center' }}
          >
            <Space size="small">
              <DesktopOutlined className="primary-color-fg" />
              <Text strong>
                {currentUsage.cpu}/{quotaLimits.cpu}
              </Text>
              <Text type="secondary" style={{ fontSize: '12px' }}>
                CPU cores
              </Text>
            </Space>
          </div>
        </Col>

        <Col xs={24} md={8} className='md:h-full'>
          <div
            className="quota-metric"
            style={{ height: '100%', display: 'flex', alignItems: 'center' }}
          >
            <Space size="small">
              <DatabaseOutlined className="success-color-fg" />
              <Text strong>
                {currentUsage.memory.toFixed(1)}/{quotaLimits.memory.toFixed(1)}
              </Text>
              <Text type="secondary" style={{ fontSize: '12px' }}>
                RAM GB
              </Text>
            </Space>
          </div>
        </Col>

        <Col xs={24} md={8} className='md:h-full'>
          <div
            className="quota-metric"
            style={{ height: '100%', display: 'flex', alignItems: 'center' }}
          >
            <Space size="small">
              <CloudOutlined className="warning-color-fg" />
              <Text strong>
                {currentUsage.instances}/{quotaLimits.instances}
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
