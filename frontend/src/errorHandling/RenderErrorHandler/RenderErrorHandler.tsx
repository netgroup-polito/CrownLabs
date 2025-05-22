import type { FC } from 'react';
import { useContext } from 'react';
import { Button } from 'antd';
import { Result } from 'antd';
import { ErrorContext } from '../ErrorContext';
import { ErrorItem } from '../ErrorHandler';
import type { CustomError } from '../utils';

export interface IRenderErrorHandlerProps {
  className?: string;
  errors: CustomError[];
}

const RenderErrorHandler: FC<IRenderErrorHandlerProps> = ({ ...props }) => {
  const { className, errors } = props;
  const { flushRenderError } = useContext(ErrorContext);
  const errorsMapped = errors.map((e, i) => <ErrorItem key={i} item={e} />);
  return (
    <div
      className={`flex flex-col h-full w-full justify-center items-center ${className}`}
    >
      <div className="flex w-full justify-center">
        <Result
          className="p-0"
          //It's not a real 500 error, that's just for visualize a cute image with the component
          status="500"
          title="Application Error."
          //subTitle={'Sorry, something went wrong.'}
        />
      </div>
      <div className="flex w-1/2 justify-center mt-6">{errorsMapped}</div>
      <div className="flex w-full justify-center mt-6">
        <Button
          type="primary"
          shape="round"
          size="large"
          onClick={flushRenderError}
        >
          Refresh
        </Button>
      </div>
    </div>
  );
};

export default RenderErrorHandler;
