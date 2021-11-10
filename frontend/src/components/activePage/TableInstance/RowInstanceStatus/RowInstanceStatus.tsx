import { FC } from 'react';
import { Tooltip } from 'antd';
import {
  CheckCircleOutlined,
  LoadingOutlined,
  CloseCircleOutlined,
  WarningOutlined,
  PauseCircleOutlined,
} from '@ant-design/icons';
import { VmStatus } from '../../../../utils';

export interface IRowInstanceStatusProps {
  status: VmStatus;
}

const RowInstanceStatus: FC<IRowInstanceStatusProps> = ({ ...props }) => {
  const { status } = props;

  const font20px = { fontSize: '20px' };
  const statusIcon = {
    CreationLoopBackoff: (
      <WarningOutlined className="warning-color-fg" style={font20px} />
    ),
    Running: <LoadingOutlined className="warning-color-fg" style={font20px} />,
    Importing: (
      <LoadingOutlined className="warning-color-fg" style={font20px} />
    ),
    VmiReady: (
      <CheckCircleOutlined className="success-color-fg" style={font20px} />
    ),
    Failed: (
      <CloseCircleOutlined className="danger-color-fg" style={font20px} />
    ),
    VmiOff: (
      <PauseCircleOutlined className="warning-color-fg" style={font20px} />
    ),
    Starting: <LoadingOutlined className="warning-color-fg" style={font20px} />,
    Stopping: <LoadingOutlined className="warning-color-fg" style={font20px} />,
  };

  return (
    <div className="flex gap-4 items-center">
      <Tooltip title={status}>{statusIcon[status || 'Starting']}</Tooltip>
    </div>
  );
};

export default RowInstanceStatus;
