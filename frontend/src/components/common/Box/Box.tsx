import { FC } from 'react';
import { Card } from 'antd';
import './Box.less';
import { BoxHeaderSize } from '../../../utils';
export interface IBoxProps {
  header?: BoxHeader;
  footer?: JSX.Element;
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
    large: 'h-28',
  };

  return (
    <>
      <Card
        className="flex-auto flex flex-col shadow-lg rounded-3xl cl-card-box"
        bordered={false}
      >
        <div className="inner">
          {header && (
            <div
              className={`${
                size ? classPerSize[size] : ''
              } flex-none w-full flex justify-center items-center box-header`}
            >
              {left && (
                <div className="flex-none h-full flex justify-center items-center pl-10">
                  {left}
                </div>
              )}
              {center && (
                <div className="flex-grow h-full flex justify-center items-center px-5">
                  {center}
                </div>
              )}
              {right && (
                <div className="flex-none h-full flex justify-center items-center pr-10">
                  {right}
                </div>
              )}
            </div>
          )}
        </div>
        {children}
        <div className="inner">
          {footer && (
            <div className="flex-none w-full py-10 flex justify-center">
              {footer}
            </div>
          )}
        </div>
      </Card>
    </>
  );
};

export default Box;
