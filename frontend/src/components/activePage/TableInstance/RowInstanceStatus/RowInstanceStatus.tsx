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
  environments?: Array<{
    name: string;
    phase?: Phase2;
  }>;
}

const RowInstanceStatus: FC<IRowInstanceStatusProps> = ({ ...props }) => {
  const { status, environments } = props;

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

  const getTooltipContent = () => {
    const mainStatus = findKeyByValue(Phase2, status || Phase2.Starting);
    
    if (!environments || environments.length === 0) {
      return mainStatus;
    }

    return (
      <div>
        <div><strong>Instance:</strong> {mainStatus}</div>
        {environments.length > 0 && (
          <>
            <div style={{ marginTop: '2px' }}><strong>Environments:</strong></div>
            {environments.map((env, index) => (
              <div key={index} style={{ marginLeft: '8px' }}>
                {env.name}: {findKeyByValue(Phase2, env.phase || Phase2.Starting)}
              </div>
            ))}
          </>
        )}
      </div>
    );
  };

  return (
    <div className="flex gap-4 items-center">
      <Tooltip title={getTooltipContent()}>
        {statusIcon[status || Phase2.Starting]}
      </Tooltip>
    </div>
  );
};

export default RowInstanceStatus;
