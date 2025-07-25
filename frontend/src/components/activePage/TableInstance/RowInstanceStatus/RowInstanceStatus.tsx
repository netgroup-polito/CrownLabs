import type { FC } from 'react';
import { Tooltip } from 'antd';
import {
  CheckCircleOutlined,
  LoadingOutlined,
  CloseCircleOutlined,
  WarningOutlined,
  PauseCircleOutlined,
} from '@ant-design/icons';
import { Phase } from '../../../../generated-types';
import { findKeyByValue } from '../../../../utils';

export interface IRowInstanceStatusProps {
  status: Phase;
}

const RowInstanceStatus: FC<IRowInstanceStatusProps> = ({ ...props }) => {
  const { status } = props;

  const font20px = { fontSize: '20px' };
  const statusIcon = {
    [Phase.Empty]: (
      <WarningOutlined className="warning-color-fg" style={font20px} />
    ),
    [Phase.CreationLoopBackoff]: (
      <WarningOutlined className="warning-color-fg" style={font20px} />
    ),
    [Phase.Running]: (
      <LoadingOutlined className="warning-color-fg" style={font20px} />
    ),
    [Phase.Importing]: (
      <LoadingOutlined className="warning-color-fg" style={font20px} />
    ),
    [Phase.ResourceQuotaExceeded]: (
      <CloseCircleOutlined className="danger-color-fg" style={font20px} />
    ),
    [Phase.Ready]: (
      <CheckCircleOutlined className="success-color-fg" style={font20px} />
    ),
    [Phase.Failed]: (
      <CloseCircleOutlined className="danger-color-fg" style={font20px} />
    ),
    [Phase.Off]: (
      <PauseCircleOutlined className="warning-color-fg" style={font20px} />
    ),
    [Phase.Starting]: (
      <LoadingOutlined className="warning-color-fg" style={font20px} />
    ),
    [Phase.Stopping]: (
      <LoadingOutlined className="warning-color-fg" style={font20px} />
    ),
  };

  return (
    <div className="flex gap-4 items-center">
      <Tooltip title={findKeyByValue(Phase, status || Phase.Starting)}>
        {statusIcon[status || Phase.Starting]}
      </Tooltip>
    </div>
  );
};

export default RowInstanceStatus;
