import { Row, Col, Typography, Space, Progress } from 'antd';
import {
  DesktopOutlined,
  CloudOutlined,
  DatabaseOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons';
import type { FC } from 'react';
import { useContext, useMemo } from 'react';
import { TenantContext } from '../../../contexts/TenantContext';
import { ItPolitoCrownlabsV1alpha2Instance } from '../../../generated-types';
import './QuotaDisplay.less';

const { Text, Title } = Typography;

export interface IQuotaDisplayProps {
  tenantNamespace: string;
  instances: ItPolitoCrownlabsV1alpha2Instance[];
  templates: any[]; // keep as is
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
  tenantNamespace,
  templates,
  instances,
  workspaceQuota,
}) => {
  // Use workspaceQuota directly
  const quota = workspaceQuota;

  // Calculate current usage: sum resources for each template times its running instances
  const currentUsage = useMemo(() => {
    let usedCpu = 0;
    let usedMemory = 0;
    let runningInstances = 0;

    templates.forEach(template => {
      const count = template.instances?.length || 0;
      runningInstances += count;
      if (template.resources) {
        usedCpu += (template.resources.cpu || 0) * count;
        usedMemory += parseMemory(template.resources.memory || '0') * count;
      }
    });

    return {
      cpu: usedCpu,
      memory: usedMemory,
      instances: runningInstances,
    };
  }, [templates]);

  // Quota limits with defaults
  const quotaLimits = {
    cpu: quota?.cpu ? parseInt(quota.cpu) : 8,
    memory: quota?.memory ? parseMemory(quota.memory) : 16,
    instances: quota?.instances || 8,
  };

  // Calculate percentages
  const cpuPercent = Math.round((currentUsage.cpu / quotaLimits.cpu) * 100);
  const memoryPercent = Math.round(
    (currentUsage.memory / quotaLimits.memory) * 100
  );
  const instancesPercent = Math.round(
    (currentUsage.instances / quotaLimits.instances) * 100
  );

  const getProgressColor = (percent: number) => {
    if (percent > 80) return '#ff4d4f';
    if (percent > 60) return '#faad14';
    return '#52c41a';
  };

  return (
    <div className="quota-display-container">
      <Row align="middle" justify="space-between">
        <Col xs={24} sm={6}>
          <Space direction="vertical" size="small">
            <Space>
              <InfoCircleOutlined className="primary-color-fg" />
              <Title level={5} style={{ margin: 0 }}>
                Resource Usage
              </Title>
            </Space>
            <Text type="secondary" style={{ fontSize: '12px' }}>
              Current usage from running instances
            </Text>
          </Space>
        </Col>

        <Col xs={24} sm={18}>
          <Row gutter={[24, 16]} justify="end">
            <Col xs={24} sm={8}>
              <div className="quota-metric">
                <Space
                  direction="vertical"
                  size="small"
                  style={{ width: '100%' }}
                >
                  <Space>
                    <DesktopOutlined className="primary-color-fg" />
                    <Text strong>
                      {currentUsage.cpu}/{quotaLimits.cpu}
                    </Text>
                    <Text type="secondary" style={{ fontSize: '12px' }}>
                      CPU cores
                    </Text>
                  </Space>
                  <Progress
                    percent={cpuPercent}
                    size="small"
                    strokeColor={getProgressColor(cpuPercent)}
                    showInfo={false}
                  />
                  <Text type="secondary" style={{ fontSize: '11px' }}>
                    {cpuPercent}% used
                  </Text>
                </Space>
              </div>
            </Col>

            <Col xs={24} sm={8}>
              <div className="quota-metric">
                <Space
                  direction="vertical"
                  size="small"
                  style={{ width: '100%' }}
                >
                  <Space>
                    <DatabaseOutlined className="success-color-fg" />
                    <Text strong>
                      {currentUsage.memory.toFixed(1)}/{quotaLimits.memory}
                    </Text>
                    <Text type="secondary" style={{ fontSize: '12px' }}>
                      GB RAM
                    </Text>
                  </Space>
                  <Progress
                    percent={memoryPercent}
                    size="small"
                    strokeColor={getProgressColor(memoryPercent)}
                    showInfo={false}
                  />
                  <Text type="secondary" style={{ fontSize: '11px' }}>
                    {memoryPercent}% used
                  </Text>
                </Space>
              </div>
            </Col>

            <Col xs={24} sm={8}>
              <div className="quota-metric">
                <Space
                  direction="vertical"
                  size="small"
                  style={{ width: '100%' }}
                >
                  <Space>
                    <CloudOutlined className="warning-color-fg" />
                    <Text strong>
                      {currentUsage.instances}/{quotaLimits.instances}
                    </Text>
                    <Text type="secondary" style={{ fontSize: '12px' }}>
                      Running instances
                    </Text>
                  </Space>
                  <Progress
                    percent={instancesPercent}
                    size="small"
                    strokeColor={getProgressColor(instancesPercent)}
                    showInfo={false}
                  />
                  <Text type="secondary" style={{ fontSize: '11px' }}>
                    {instancesPercent}% used
                  </Text>
                </Space>
              </div>
            </Col>
          </Row>
        </Col>
      </Row>
    </div>
  );
};

export default QuotaDisplay;
