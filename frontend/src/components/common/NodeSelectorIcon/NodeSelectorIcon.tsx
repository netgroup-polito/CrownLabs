import type { FC } from 'react';
import { Tooltip } from 'antd';

import { AimOutlined } from '@ant-design/icons';
import { cleanupLabels } from '../../../utils';

export interface INodeSelectorIconProps {
  nodeSelector: JSON;
  isOnWorkspace: boolean;
}

const NodeSelectorIcon: FC<INodeSelectorIconProps> = ({ ...props }) => {
  const { nodeSelector, isOnWorkspace } = props;

  const displaySel = Object.entries(nodeSelector)
    .map(([k, v]) => `${cleanupLabels(k)}=${v}`)
    .join(',');

  const tooltipText = !isOnWorkspace
    ? `This instance started on a node with ${displaySel}`
    : displaySel
      ? `This instance will be started on nodes with ${displaySel}`
      : 'This instance can be started choosing the target node';

  return (
    <Tooltip title={<div className="text-center">{tooltipText}</div>}>
      <div className="primary-color-fg flex items-center">
        <AimOutlined width="22px" />
      </div>
    </Tooltip>
  );
};

export default NodeSelectorIcon;
