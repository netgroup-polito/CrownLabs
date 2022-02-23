import { FC } from 'react';
import { Tooltip } from 'antd';
import {
  CheckCircleOutlined,
  LoadingOutlined,
  CloseCircleOutlined,
  WarningOutlined,
  PauseCircleOutlined,
} from '@ant-design/icons';
import { Phase } from '../../../../generated-types';

export interface IRowInstanceStatusProps {
  status: Phase;
}

const RowInstanceStatus: FC<IRowInstanceStatusProps> = ({ ...props }) => {
  const { status } = props;

  const font20px = { fontSize: '20px' };
  const statusIcon = {
    Unknown: <WarningOutlined className="warning-color-fg" style={font20px} />,
    CreationLoopBackoff: (
      <WarningOutlined className="warning-color-fg" style={font20px} />
    ),
    Running: <LoadingOutlined className="warning-color-fg" style={font20px} />,
    Importing: (
      <LoadingOutlined className="warning-color-fg" style={font20px} />
    ),
    ResourceQuotaExceeded: (
      <CloseCircleOutlined className="danger-color-fg" style={font20px} />
    ),
    Ready: (
      <CheckCircleOutlined className="success-color-fg" style={font20px} />
    ),
    Failed: (
      <CloseCircleOutlined className="danger-color-fg" style={font20px} />
    ),
    Off: <PauseCircleOutlined className="warning-color-fg" style={font20px} />,
    Starting: <LoadingOutlined className="warning-color-fg" style={font20px} />,
    Stopping: <LoadingOutlined className="warning-color-fg" style={font20px} />,
  };

  return (
    <div className="flex gap-4 items-center">
      <Tooltip title={status || Phase.Starting}>
        {statusIcon[status || Phase.Starting]}
      </Tooltip>
    </div>
  );
};

export default RowInstanceStatus;
