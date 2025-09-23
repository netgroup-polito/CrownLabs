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
  const { header, children, footer } = props;
  const { center, left, right, size } = header || {};

  const classPerSize = {
    small: 'h-14',
    middle: 'h-20',
    large: 'h-24',
  };

  return (
    <>
      <Card
        className="flex-auto flex flex-col shadow-lg rounded-3xl cl-card-box h-full"
        styles={{
          body: { height: '100%', flexDirection: 'column', display: 'flex' },
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
        <div className="w-full flex-grow">{children}</div>
        <div className="w-full flex-none inner">{footer}</div>
      </Card>
    </>
  );
};

export default Box;
