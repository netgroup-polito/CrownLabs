import type { FC } from 'react';
import { Tooltip } from 'antd';
import {
  CheckCircleOutlined,
  LoadingOutlined,
  CloseCircleOutlined,
  StopOutlined,
} from '@ant-design/icons';
import { Phase5 } from '../../../../generated-types';
import { findKeyByValue } from '../../../../utils';

export interface IRowShVolStatusProps {
  status: Phase5;
}

const RowShVolStatus: FC<IRowShVolStatusProps> = ({ ...props }) => {
  const { status } = props;

  const font20px = { fontSize: '20px' };
  const statusIcon = {
    [Phase5.Empty]: (
      <LoadingOutlined className="warning-color-fg" style={font20px} />
    ),
    [Phase5.Pending]: (
      <LoadingOutlined className="warning-color-fg" style={font20px} />
    ),
    [Phase5.Provisioning]: (
      <LoadingOutlined className="warning-color-fg" style={font20px} />
    ),
    [Phase5.Ready]: (
      <CheckCircleOutlined className="success-color-fg" style={font20px} />
    ),
    [Phase5.Deleting]: (
      <StopOutlined className="danger-color-fg" style={font20px} />
    ),
    [Phase5.ResourceQuotaExceeded]: (
      <CloseCircleOutlined className="danger-color-fg" style={font20px} />
    ),
    [Phase5.Error]: (
      <CloseCircleOutlined className="danger-color-fg" style={font20px} />
    ),
  };

  return (
    <div className="flex gap-4 items-center">
      <Tooltip title={findKeyByValue(Phase5, status || Phase5.Pending)}>
        {statusIcon[status || Phase5.Pending]}
      </Tooltip>
    </div>
  );
};

export default RowShVolStatus;
