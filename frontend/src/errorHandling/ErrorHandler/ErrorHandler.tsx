import { CloseOutlined } from '@ant-design/icons';
import { Modal, Typography } from 'antd';
import { Button } from 'antd';
import type { FC } from 'react';
import { ErrorItem } from '.';
import type { CustomError } from '../utils';
import { ErrorTypes } from '../utils';

const { Text } = Typography;

export interface IErrorHandlerProps {
  errorsQueue: Array<CustomError>;
  show: boolean;
  dismiss: () => void;
}

const ErrorHandler: FC<IErrorHandlerProps> = ({ ...props }) => {
  const { errorsQueue, dismiss, show } = props;

  const titleFromType = {
    [ErrorTypes.ApolloError]: 'Server Error',
    [ErrorTypes.AuthError]: 'Authentication Error',
    [ErrorTypes.RenderError]: 'Application Error',
    [ErrorTypes.GenericError]: 'Generic Error',
  };

  return (
    <>
      <Modal
        footer={false}
        centered
        title={
          errorsQueue.length ? (
            <Text type="danger" strong className="text-2xl">
              {titleFromType[errorsQueue[0].getType()]}
            </Text>
          ) : null
        }
        open={show}
        closable={false}
        width={800}
      >
        <div className="flex-column justify-start w-full items-center p-4">
          {errorsQueue.length > 0 && (
            <div className="flex-column items-center w-full">
              <ErrorItem item={errorsQueue[0]} />
            </div>
          )}
        </div>
        <div className="flex justify-center mt-6">
          <div className="h-full flex justify-center items-center gap-8">
            <Button
              size="large"
              shape="round"
              icon={<CloseOutlined />}
              color="red"
              onClick={dismiss}
            >
              Dismiss
            </Button>
          </div>
        </div>
      </Modal>
    </>
  );
};

export default ErrorHandler;
