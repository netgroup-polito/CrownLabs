import type { FC } from 'react';
import { Card } from 'antd';
import './Box.less';
import type { BoxHeaderSize } from '../../../utils';
export interface IBoxProps {
  header?: BoxHeader;
  footer?: React.ReactNode;
  children?: React.ReactNode;
}

export type BoxHeader = {
  left?: React.ReactNode;
  right?: React.ReactNode;
  center?: React.ReactNode;
  size: BoxHeaderSize;
};

const Box: FC<IBoxProps> = ({ ...props }) => {
  const { header, children } = props;
  const { center, left, right, size } = header || {};

  const classPerSize = {
    small: 'h-14',
    middle: 'h-20',
    large: 'h-24',
  };

  return (
    <Card
      className="flex-auto flex flex-col shadow-lg rounded-3xl cl-card-box h-full"
      // make the card body a true flex item so inner scroll surfaces can compute height
      bodyStyle={{
        display: 'flex',
        flexDirection: 'column',
        flex: '1 1 auto',
        minHeight: 0,
      }}
    >
      <div className="w-full flex-none">
        {header && (
          <div
            className={`${
              size ? classPerSize[size] : ''
            } flex justify-center items-center box-header`}
          >
            <div className="flex-none h-full">{left}</div>
            <div className="flex-grow h-full">{center}</div>
            <div className="flex-none h-full">{right}</div>
          </div>
        )}
      </div>
      {/* allow Box body to shrink inside the flex chain */}
      <div className="w-full flex-grow min-h-0" style={{ display: 'flex' }}>
        {children}
      </div>
    </Card>
  );
};

export default Box;
