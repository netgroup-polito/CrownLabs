import { Alert, Space, Typography } from 'antd';
import Button from 'antd-button-color';
import { ApolloError } from '@apollo/client';
import { KeycloakError } from 'keycloak-js';
import { FC, useState } from 'react';
import { CustomError, ErrorTypes } from '../utils';

const { Text } = Typography;

export interface IErrorItemProps {
  item: CustomError;
}

const ErrorItem: FC<IErrorItemProps> = ({ ...props }) => {
  const { item } = props;
  const [expanded, setExpanded] = useState(false);

  const errorDesc = (
    <>
      {expanded && (
        <Text className="mt-4 block max-h-72 overflow-auto" copyable keyboard>
          {item.getType() === ErrorTypes.KeycloakError
            ? (item.getError() as KeycloakError).error_description
            : (item.getError() as ApolloError | Error).stack}
        </Text>
      )}
    </>
  );

  const messageFromType = {
    [ErrorTypes.ApolloError]: (item.getError() as ApolloError).message,
    [ErrorTypes.KeycloakError]: (item.getError() as KeycloakError).error,
    [ErrorTypes.RenderError]: (
      <>
        <div>
          The CrownLabs web client encountered an error that cannot be solved
          automatically.
        </div>
        <div>Contact support if this keeps occurring.</div>
      </>
    ),
    [ErrorTypes.GenericError]: (item.getError() as ApolloError).message,
  };

  return (
    <div className="flex w-full justify-center mb-2">
      <Alert
        className="w-full"
        message={
          <>
            <Text strong className="block">
              {messageFromType[item.getType()]}
            </Text>
          </>
        }
        showIcon
        description={errorDesc}
        type="error"
        action={
          <Space
            direction="vertical"
            className="flex justify-center items-center h-full"
          >
            <Button
              size="small"
              danger
              type="ghost"
              onClick={() => setExpanded(old => !old)}
            >
              {expanded ? 'Hide debug info' : 'Show debug info'}
            </Button>
          </Space>
        }
      />
    </div>
  );
};

export default ErrorItem;
