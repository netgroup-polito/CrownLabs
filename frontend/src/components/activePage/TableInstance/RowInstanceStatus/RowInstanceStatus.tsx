import type { FC } from 'react';
import { Tooltip } from 'antd';
import {
  CheckCircleOutlined,
  LoadingOutlined,
  CloseCircleOutlined,
  WarningOutlined,
  PauseCircleOutlined,
} from '@ant-design/icons';
import { Phase2 } from '../../../../generated-types';
import { findKeyByValue } from '../../../../utils';

export interface IRowInstanceStatusProps {
  status: Phase2;
}

const RowInstanceStatus: FC<IRowInstanceStatusProps> = ({ ...props }) => {
  const { status } = props;

  const font20px = { fontSize: '20px' };
  const statusIcon = {
    [Phase2.Empty]: (
      <WarningOutlined className="warning-color-fg" style={font20px} />
    ),
    [Phase2.CreationLoopBackoff]: (
      <WarningOutlined className="warning-color-fg" style={font20px} />
    ),
    [Phase2.Running]: (
      <LoadingOutlined className="warning-color-fg" style={font20px} />
    ),
    [Phase2.Importing]: (
      <LoadingOutlined className="warning-color-fg" style={font20px} />
    ),
    [Phase2.ResourceQuotaExceeded]: (
      <CloseCircleOutlined className="danger-color-fg" style={font20px} />
    ),
    [Phase2.Ready]: (
      <CheckCircleOutlined className="success-color-fg" style={font20px} />
    ),
    [Phase2.Failed]: (
      <CloseCircleOutlined className="danger-color-fg" style={font20px} />
    ),
    [Phase2.Off]: (
      <PauseCircleOutlined className="warning-color-fg" style={font20px} />
    ),
    [Phase2.Starting]: (
      <LoadingOutlined className="warning-color-fg" style={font20px} />
    ),
    [Phase2.Stopping]: (
      <LoadingOutlined className="warning-color-fg" style={font20px} />
    ),
  };

  return (
    <div className="flex gap-4 items-center">
      <Tooltip title={findKeyByValue(Phase2, status || Phase2.Starting)}>
        {statusIcon[status || Phase2.Starting]}
      </Tooltip>
    </div>
  );
};

export default RowInstanceStatus;
