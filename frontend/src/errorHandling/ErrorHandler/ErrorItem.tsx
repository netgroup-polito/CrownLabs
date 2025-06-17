import { Alert, Space, Typography } from 'antd';
import { Button } from 'antd';
import type { ApolloError } from '@apollo/client';
import type { FC } from 'react';
import { useState } from 'react';
import type { CustomError } from '../utils';
import { ErrorTypes } from '../utils';
import type { ErrorContext } from 'react-oidc-context';

const { Text } = Typography;

export interface IErrorItemProps {
  item: CustomError;
}

const ErrorItem: FC<IErrorItemProps> = ({ ...props }) => {
  const { item } = props;
  const [expanded, setExpanded] = useState(false);

  const out = (() => {
    const typ = item.getType();
    const err = item.getError();

    if (typ === ErrorTypes.ApolloError || err.name === 'ApolloError') {
      const ae = err as ApolloError;
      return {
        msg:
          err.message === 'Failed to fetch' ? 'Connection error' : err.message,
        more: (
          <ul>
            {ae.clientErrors?.length ? (
              <li>Client errs: {JSON.stringify(ae.clientErrors)}</li>
            ) : null}
            {ae.extraInfo && <li>Extra: {ae.extraInfo}</li>}
            {ae.networkError && (
              <li>Network error: {ae.networkError?.message}</li>
            )}
            {ae.protocolErrors?.length ? (
              <li>
                Protocol errors:{' '}
                <ul>{ae.protocolErrors?.map(e => <li>{e.message}</li>)}</ul>
              </li>
            ) : null}
            {ae.graphQLErrors.length ? (
              <li>
                GraphQL Errors:
                {ae.graphQLErrors.map(e => (
                  <li>
                    {e.message}
                    <ul>
                      <li>Path: {e.path?.join('/')}</li>
                      <li>
                        Locs: {e.locations?.map(l => l.line + '@' + l.column)}
                      </li>
                    </ul>
                  </li>
                ))}
              </li>
            ) : null}
            {Object.keys(ae.cause || {}).length ? (
              <li>
                Cause: <pre>{JSON.stringify(ae.cause, null, 2)}</pre>
              </li>
            ) : null}
          </ul>
        ),
      };
    }

    if (typ === ErrorTypes.RenderError) {
      return {
        msg: (
          <>
            <div>
              The CrownLabs web client encountered an error that cannot be
              solved automatically.
            </div>
            <div>Contact support if this keeps occurring.</div>
          </>
        ),
        more: <pre>{err.stack}</pre>,
      };
    }

    return {
      msg: err.message,
      more: JSON.stringify(err as ErrorContext),
    };
  })();

  return (
    <div className="flex w-full justify-center mb-2">
      <Alert
        className="w-full"
        message={
          <>
            <Text strong className="block">
              {out.msg}
            </Text>
          </>
        }
        showIcon
        description={
          expanded && (
            <Text className="mt-4 block max-h-72 overflow-auto" copyable>
              {out.more}
            </Text>
          )
        }
        type="error"
        action={
          <Space
            direction="vertical"
            className="flex justify-center items-center h-full"
          >
            <Button
              size="small"
              danger
              type="link"
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
