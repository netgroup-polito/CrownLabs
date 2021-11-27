import { FC, useContext } from 'react';
import Button from 'antd-button-color';
import { Result } from 'antd';
import { ErrorContext } from '../ErrorContext';

export interface IRenderErrorAlertProps {
  className?: string;
}

const RenderErrorAlert: FC<IRenderErrorAlertProps> = ({ ...props }) => {
  const { className } = props;
  const { flushRenderError } = useContext(ErrorContext);
  return (
    <div className={`flex flex-col h-full w-full justify-center ${className}`}>
      <div className="flex w-full justify-center">
        <Result
          className="p-0"
          //It's not a real 500 error, that's just for visualize a cute image with the component
          status="500"
          title="Application error."
          subTitle={'Sorry, something went wrong.'}
        />
      </div>
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

export default RenderErrorAlert;
