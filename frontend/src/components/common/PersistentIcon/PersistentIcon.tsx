import { FC } from 'react';
import { Tooltip } from 'antd';

import { ReactComponent as SvgInfinite } from '../../../assets/infinite.svg';

export interface IPersistentIconProps {}

const PersistentIcon: FC<IPersistentIconProps> = ({ ...props }) => {
  return (
    <Tooltip
      title={
        <>
          <div className="text-center">
            These Instances can be stopped and restarted without being deleted.
          </div>
          <div className="text-center">
            Your files won't be deleted in case of an internal disservice of
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
