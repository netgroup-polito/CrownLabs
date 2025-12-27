import { type FC, useContext, useMemo, useState } from 'react';
import { Tooltip } from 'antd';
import { Button } from 'antd';
import {
  ExclamationCircleOutlined,
  PauseCircleOutlined,
  PlayCircleOutlined,
} from '@ant-design/icons';
import { type Instance } from '../../../../utils';
import { Phase2, useApplyInstanceMutation } from '../../../../generated-types';
import { setInstanceRunning } from '../../../../utilsLogic';
import { ErrorContext } from '../../../../errorHandling/ErrorContext';
import type { ApolloError } from '@apollo/client';
import type { IQuota } from '../../../../contexts/OwnedInstancesContext';
import { cn } from '../../../../utils/style';

export interface IRowInstanceActionsPersistentProps {
  extended: boolean;
  instance: Instance;
  workspaceAvailableQuota: IQuota;
}

const RowInstanceActionsPersistent: FC<IRowInstanceActionsPersistentProps> = ({
  extended,
  instance,
  workspaceAvailableQuota,
}) => {
  const { status } = instance;

  const font22px = { fontSize: '22px' };

  const [disabled, setDisabled] = useState(false);

  const { apolloErrorCatcher } = useContext(ErrorContext);
  const [applyInstanceMutation] = useApplyInstanceMutation({
    onError: apolloErrorCatcher,
  });

  const mutateInstanceStatus = async (running: boolean) => {
    if (!disabled) {
      setDisabled(true);
      try {
        const result = await setInstanceRunning(
          running,
          instance,
          applyInstanceMutation,
        );
        if (result) setTimeout(setDisabled, 400, false);
      } catch (err) {
        apolloErrorCatcher(err as ApolloError);
      }
    }
  };

  const canStartInstance = useMemo(() => {
    if (!workspaceAvailableQuota) return true;

    return (
      workspaceAvailableQuota.instances >= 1 &&
      workspaceAvailableQuota.cpu >= instance.resources.cpu &&
      workspaceAvailableQuota.memory >= instance.resources.memory
      // TODO: add this when disk quota is available - workspaceAvailableQuota.disk >= instance.resources.disk
    );
  }, [instance, workspaceAvailableQuota]);

  return status === Phase2.Ready || status === Phase2.ResourceQuotaExceeded ? (
    <Tooltip placement="top" title="Pause">
      <Button
        loading={disabled}
        className={`hidden ${
          extended ? 'sm:block' : 'xs:block'
        } flex items-center`}
        color="orange"
        type="link"
        shape="circle"
        size="middle"
        disabled={disabled}
        icon={
          <PauseCircleOutlined
            className="flex justify-center items-center"
            style={font22px}
          />
        }
        onClick={() => mutateInstanceStatus(false)}
      />
    </Tooltip>
  ) : status === Phase2.Off ? (
    <Tooltip placement="top" title="Start">
      <Button
        loading={disabled}
        className={`hidden ${extended ? 'sm:block' : 'xs:block'} py-0`}
        type="link"
        shape="circle"
        size="middle"
        disabled={disabled || !canStartInstance}
        icon={
          <PlayCircleOutlined
            className={cn(
              'flex justify-center items-center',
              canStartInstance && 'success-color-fg',
            )}
            style={font22px}
          />
        }
        onClick={() => mutateInstanceStatus(true)}
      />
    </Tooltip>
  ) : (
    <Tooltip placement="top" title={'Current instance Status: ' + status}>
      <div className="cursor-not-allowed">
        <Button
          className={`hidden pointer-events-none ${
            extended ? 'sm:block' : 'xs:block'
          } py-0`}
          color="primary"
          type="link"
          shape="circle"
          size="middle"
          disabled={true}
          icon={
            <ExclamationCircleOutlined
              className="flex justify-center items-center"
              style={font22px}
            />
          }
        />
      </div>
    </Tooltip>
  );
};

export default RowInstanceActionsPersistent;
