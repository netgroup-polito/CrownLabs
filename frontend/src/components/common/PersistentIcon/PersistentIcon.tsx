import type { FC } from 'react';
import { Tooltip } from 'antd';

import SvgInfinite from '../../../assets/infinite.svg?react';

const PersistentIcon: FC = () => {
  return (
    <Tooltip
      title={
        <>
          <div className="text-center">
            This instance can be stopped and restarted without being deleted.
          </div>
          <div className="text-center">
            Your files will be preserved also in case of a malfunctioning of
            CrownLabs.
          </div>
        </>
      }
    >
      <div className="success-color-fg flex items-center">
        <SvgInfinite width="22px" />
      </div>
    </Tooltip>
  );
};

export default PersistentIcon;
