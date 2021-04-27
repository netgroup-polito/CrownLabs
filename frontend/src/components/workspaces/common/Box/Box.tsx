import { FC } from 'react';
import { Card } from 'antd';
import '../../../../index.less'; //To delete, usefull only to storybook

export interface IBoxProps {
  headLeft?: React.ReactNode;
  headRight?: React.ReactNode;
  headTitle?: string;
  footer?: React.ReactNode;
}

const Box: FC<IBoxProps> = ({ ...props }) => {
  const { headLeft, headRight, headTitle, children, footer } = props;
  return (
    <>
      <Card
        className="flex-auto flex flex-col shadow-lg rounded-3xl"
        bordered={false}
        bodyStyle={{ padding: '0px' }}
      >
        {headLeft || headRight || headTitle ? (
          <div className="w-full h-28 flex-none flex justify-center items-center border-t-0 border-l-0 border-r-0 border-b-3 border-solid border-black">
            {headLeft ? (
              <div className="flex-none h-28 flex justify-center items-center pl-10">
                {headLeft}
              </div>
            ) : (
              ''
            )}
            {headTitle ? (
              <div className="flex-grow h-28 flex justify-center items-center px-5">
                <p className="md:text-4xl text-2xl text-center mb-0">
                  <b>{headTitle}</b>
                </p>
              </div>
            ) : (
              ''
            )}
            {headRight ? (
              <div className="flex-none h-28 flex justify-center items-center pr-10">
                {headRight}
              </div>
            ) : (
              ''
            )}
          </div>
        ) : (
          ''
        )}
        {children}
        {footer ? (
          <div className="flex-none w-full py-10 flex justify-center">
            {footer}
          </div>
        ) : (
          ''
        )}
      </Card>
    </>
  );
};

export default Box;
