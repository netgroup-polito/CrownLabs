import { Row, Col, Typography, Space } from 'antd';
import {
  DesktopOutlined,
  CloudOutlined,
  DatabaseOutlined,
} from '@ant-design/icons';
import type { FC } from 'react';
import { useMemo } from 'react';
import './QuotaDisplay.less';

const { Text } = Typography;

export interface IQuotaDisplayProps {
  consumedQuota: {
    cpu?: string | number;
    memory?: string;
    instances?: number;
  } | null; // Allow null
  workspaceQuota: {
    cpu?: string | number;
    memory?: string;
    instances?: number;
  };
}

// Helper function to parse memory string (e.g., "4Gi" -> 4)
const parseMemory = (memoryStr: string): number => {
  if (!memoryStr) return 0;

  const match = memoryStr.match(/^(\d+(?:\.\d+)?)(.*)?$/);
  if (!match) return 0;

  const value = parseFloat(match[1]);
  const unit = match[2]?.toLowerCase() || '';

  switch (unit) {
    case 'gi':
    case 'g':
      return value;
    case 'mi':
    case 'm':
      return value / 1024;
    case 'ki':
    case 'k':
      return value / (1024 * 1024);
    case 'ti':
    case 't':
      return value * 1024;
    default:
      // Assume GB if no unit
      return value;
  }
};

const QuotaDisplay: FC<IQuotaDisplayProps> = ({
  consumedQuota,
  workspaceQuota,
}) => {
  // Use workspaceQuota directly
  const quota = workspaceQuota;

  // Calculate current usage: sum resources for each template times its running instances
  const currentUsage = useMemo(() => {
    let usedCpu = 0;
    let usedMemory = 0;
    let runningInstances = 0;

    // Calculate current usage based on consumed resources
    if (consumedQuota) {
      usedCpu =
        typeof consumedQuota.cpu === 'string'
          ? parseFloat(consumedQuota.cpu) || 0
          : consumedQuota.cpu || 0;
      usedMemory = parseMemory(consumedQuota.memory || '0');
      runningInstances = consumedQuota.instances || 0;
    }

    return {
      cpu: usedCpu,
      memory: usedMemory,
      instances: runningInstances,
    };
  }, [consumedQuota]);

  // Quota limits with defaults
  const quotaLimits = {
    cpu: quota?.cpu ? parseInt(String(quota.cpu)) : 8,
    memory: quota?.memory ? parseMemory(quota.memory) : 16,
    instances: quota?.instances || 8,
  };

  return (
    <div
      className="quota-display-container"
      style={{ height: '40px', overflow: 'hidden' }}
    >
      <Row gutter={[16, 0]} style={{ height: '100%' }}>
        <Col xs={24} sm={8} style={{ height: '100%' }}>
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

        <Col xs={24} sm={8} style={{ height: '100%' }}>
          <div
            className="quota-metric"
            style={{ height: '100%', display: 'flex', alignItems: 'center' }}
          >
            <Space size="small">
              <DatabaseOutlined className="success-color-fg" />
              <Text strong>
                {currentUsage.memory.toFixed(1)}/{quotaLimits.memory}
              </Text>
              <Text type="secondary" style={{ fontSize: '12px' }}>
                RAM GB
              </Text>
            </Space>
          </div>
        </Col>

        <Col xs={24} sm={8} style={{ height: '100%' }}>
          <div
            className="quota-metric"
            style={{ height: '100%', display: 'flex', alignItems: 'right' }}
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
