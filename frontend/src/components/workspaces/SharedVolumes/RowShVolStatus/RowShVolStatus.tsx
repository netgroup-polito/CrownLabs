import type { FC } from 'react';
import { Tooltip } from 'antd';
import {
  CheckCircleOutlined,
  LoadingOutlined,
  CloseCircleOutlined,
  StopOutlined,
} from '@ant-design/icons';
import { Phase6 } from '../../../../generated-types';
import { findKeyByValue } from '../../../../utils';

export interface IRowShVolStatusProps {
  status: Phase6;
}

const RowShVolStatus: FC<IRowShVolStatusProps> = ({ ...props }) => {
  const { status } = props;

  const font20px = { fontSize: '20px' };
  const statusIcon = {
    [Phase6.Empty]: (
      <LoadingOutlined className="warning-color-fg" style={font20px} />
    ),
    [Phase6.Pending]: (
      <LoadingOutlined className="warning-color-fg" style={font20px} />
    ),
    [Phase6.Provisioning]: (
      <LoadingOutlined className="warning-color-fg" style={font20px} />
    ),
    [Phase6.Ready]: (
      <CheckCircleOutlined className="success-color-fg" style={font20px} />
    ),
    [Phase6.Deleting]: (
      <StopOutlined className="danger-color-fg" style={font20px} />
    ),
    [Phase6.ResourceQuotaExceeded]: (
      <CloseCircleOutlined className="danger-color-fg" style={font20px} />
    ),
    [Phase6.Error]: (
      <CloseCircleOutlined className="danger-color-fg" style={font20px} />
    ),
  };

  return (
    <div className="flex gap-4 items-center">
      <Tooltip title={findKeyByValue(Phase6, status || Phase6.Pending)}>
        {statusIcon[status || Phase6.Pending]}
      </Tooltip>
    </div>
  );
};

export default RowShVolStatus;
