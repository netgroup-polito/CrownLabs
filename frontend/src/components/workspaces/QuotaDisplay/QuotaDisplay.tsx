import { Dropdown, Button, Typography, Space } from 'antd';
import {
  DesktopOutlined,
  CloudOutlined,
  DatabaseOutlined,
  HddOutlined,
  ThunderboltOutlined,
  AreaChartOutlined,
} from '@ant-design/icons';
import type { FC } from 'react';
import { useContext } from 'react';
import './QuotaDisplay.less';
import {
  OwnedInstancesContext,
  type IQuota,
} from '../../../contexts/OwnedInstancesContext';
import { formatExtendedResourceLabel } from '../../../utils';

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
    otherResources: {},
  };

  const workspaceTotalQuota: IQuota = totalQuota[workspaceName || ''] || {
    instances: 0,
    cpu: 0,
    memory: 0,
    disk: 0,
    otherResources: {},
  };

  // Helper function to safely process utilization percentages
  const calculatePercentage = (consumed: number, total: number): number => {
    if (!total || total === 0) {
      // If the limit is 0 but resources are consumed, treat as 100% (critical) to avoid green status
      return consumed > 0 ? 100 : 0;
    }
    return (consumed / total) * 100;
  };

  // Helper function to determine threshold status color based on resource utilization
  const getResourceStatusColor = (percentage: number): string => {
    if (percentage >= 100) return '#ff4d4f';
    if (percentage >= 80) return '#faad14';
    if (percentage >= 50) return '#fff652';
    return '#52c41a';
  };

  // Helper function to determine the hover tooltip message based on resource utilization
  const getResourceStatusMessage = (percentage: number): string => {
    if (percentage >= 100) return 'Resource usage: critical';
    if (percentage >= 80) return 'Resource usage: high';
    if (percentage >= 50) return 'Resource usage: moderate';
    return 'Resource usage: low';
  };

  // Compute standard infrastructure percentages
  const cpuPct = calculatePercentage(
    Number(workspaceConsumedQuota.cpu),
    Number(workspaceTotalQuota.cpu),
  );
  const memoryPct = calculatePercentage(
    workspaceConsumedQuota.memory,
    workspaceTotalQuota.memory,
  );
  const diskPct = calculatePercentage(
    workspaceConsumedQuota.disk,
    workspaceTotalQuota.disk,
  );
  const instancesPct = calculatePercentage(
    workspaceConsumedQuota.instances,
    workspaceTotalQuota.instances,
  );

  // Compute custom extended resources percentages dynamically
  const extendedResourcesKeys = Object.keys(
    workspaceTotalQuota.otherResources || {},
  );
  const extendedPcts = extendedResourcesKeys.map(key => {
    const consumed = Number(workspaceConsumedQuota.otherResources?.[key] ?? 0);
    const total = Number(workspaceTotalQuota.otherResources?.[key] ?? 0);
    return calculatePercentage(consumed, total);
  });

  // Evaluate the worst-case scenario metric among all available hardware quotas
  const maxPercentage = Math.max(
    cpuPct,
    memoryPct,
    diskPct,
    instancesPct,
    ...extendedPcts,
  );
  const mainStatusColor = getResourceStatusColor(maxPercentage);

  // Layout specification for individual status indicators inside the dropdown items
  const indicatorStyle = (color: string) => ({
    display: 'inline-block',
    width: '10px',
    height: '10px',
    borderRadius: '50%',
    backgroundColor: color,
    boxShadow: '0 1px 3px rgba(0,0,0,0.15)',
  });

  const verticalDividerStyle = {
    borderLeft: '1px solid #d9d9d9',
    height: '14px',
    marginLeft: '4px',
    marginRight: '4px',
  };

  const menuItems = [
    {
      key: 'cpu',
      label: (
        <Space>
          <span style={indicatorStyle(getResourceStatusColor(cpuPct))} />
          <span style={verticalDividerStyle} />
          <DesktopOutlined className="primary-color-fg" />
          <Text>
            CPU:{' '}
            <strong>
              {workspaceConsumedQuota.cpu}/{workspaceTotalQuota.cpu}
            </strong>{' '}
            ({cpuPct.toFixed(0)}%)
          </Text>
        </Space>
      ),
    },
    {
      key: 'ram',
      label: (
        <Space>
          <span style={indicatorStyle(getResourceStatusColor(memoryPct))} />
          <span style={verticalDividerStyle} />
          <DatabaseOutlined className="success-color-fg" />
          <Text>
            RAM:{' '}
            <strong>
              {workspaceConsumedQuota.memory.toFixed(1)}/
              {workspaceTotalQuota.memory.toFixed(1)}Gi
            </strong>{' '}
            ({memoryPct.toFixed(0)}%)
          </Text>
        </Space>
      ),
    },
    {
      key: 'disk',
      label: (
        <Space>
          <span style={indicatorStyle(getResourceStatusColor(diskPct))} />
          <span style={verticalDividerStyle} />
          <HddOutlined style={{ color: '#13c2c2' }} />
          <Text>
            Disk:{' '}
            <strong>
              {workspaceConsumedQuota.disk.toFixed(1)}/
              {workspaceTotalQuota.disk.toFixed(1)}Gi
            </strong>{' '}
            ({diskPct.toFixed(0)}%)
          </Text>
        </Space>
      ),
    },
    {
      key: 'instances',
      label: (
        <Space>
          <span style={indicatorStyle(getResourceStatusColor(instancesPct))} />
          <span style={verticalDividerStyle} />
          <CloudOutlined className="warning-color-fg" />
          <Text>
            Instances:{' '}
            <strong>
              {workspaceConsumedQuota.instances}/{workspaceTotalQuota.instances}
            </strong>{' '}
            ({instancesPct.toFixed(0)}%)
          </Text>
        </Space>
      ),
    },
    ...extendedResourcesKeys.map((key, idx) => {
      const consumed = workspaceConsumedQuota.otherResources?.[key] ?? 0;
      const total = workspaceTotalQuota.otherResources?.[key] ?? 0;
      const pct = calculatePercentage(Number(consumed), Number(total));
      const cleanLabel = formatExtendedResourceLabel(key);

      return {
        key: `ext-${idx}`,
        label: (
          <Space>
            <span style={indicatorStyle(getResourceStatusColor(pct))} />
            <span style={verticalDividerStyle} />
            <ThunderboltOutlined style={{ color: '#722ed1' }} />
            <Text>
              {cleanLabel}:{' '}
              <strong>
                {consumed}/{total}
              </strong>{' '}
              ({pct.toFixed(0)}%)
            </Text>
          </Space>
        ),
      };
    }),
  ];

  return (
    <div
      style={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}
    >
      <Dropdown
        menu={{ items: menuItems }}
        trigger={['click']}
        placement="bottomRight"
      >
        <Button
          shape="circle"
          size="large"
          className="quota-status-button"
          icon={<AreaChartOutlined />}
          style={{
            backgroundColor: mainStatusColor,
            borderColor: mainStatusColor,
            width: '42px',
            height: '42px',
            boxShadow: '0 2px 8px rgba(0,0,0,0.15)',
            transition: 'transform 0.2s, background-color 0.3s',
            color: '#fff',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
          title={getResourceStatusMessage(maxPercentage)}
        />
      </Dropdown>
    </div>
  );
};

export default QuotaDisplay;
