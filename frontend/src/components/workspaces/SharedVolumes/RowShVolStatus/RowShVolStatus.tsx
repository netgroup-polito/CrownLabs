import type { FC } from 'react';
import { Tooltip } from 'antd';
import {
  CheckCircleOutlined,
  LoadingOutlined,
  CloseCircleOutlined,
  StopOutlined,
} from '@ant-design/icons';
import { Phase3 } from '../../../../generated-types';
import { findKeyByValue } from '../../../../utils';

export interface IRowShVolStatusProps {
  status: Phase3;
}

const RowShVolStatus: FC<IRowShVolStatusProps> = ({ ...props }) => {
  const { status } = props;

  const font20px = { fontSize: '20px' };
  const statusIcon = {
    [Phase3.Empty]: (
      <LoadingOutlined className="warning-color-fg" style={font20px} />
    ),
    [Phase3.Pending]: (
      <LoadingOutlined className="warning-color-fg" style={font20px} />
    ),
    [Phase3.Provisioning]: (
      <LoadingOutlined className="warning-color-fg" style={font20px} />
    ),
    [Phase3.Ready]: (
      <CheckCircleOutlined className="success-color-fg" style={font20px} />
    ),
    [Phase3.Deleting]: (
      <StopOutlined className="danger-color-fg" style={font20px} />
    ),
    [Phase3.ResourceQuotaExceeded]: (
      <CloseCircleOutlined className="danger-color-fg" style={font20px} />
    ),
    [Phase3.Error]: (
      <CloseCircleOutlined className="danger-color-fg" style={font20px} />
    ),
  };

  return (
    <div className="flex gap-4 items-center">
      <Tooltip title={findKeyByValue(Phase3, status || Phase3.Pending)}>
        {statusIcon[status || Phase3.Pending]}
      </Tooltip>
    </div>
  );
};

export default RowShVolStatus;
