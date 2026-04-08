import {
  DeleteOutlined,
  ExportOutlined,
  DownOutlined,
  CodeOutlined,
} from '@ant-design/icons';
import { Tooltip, Dropdown } from 'antd';
import type { MenuProps } from 'antd';
import { Button } from 'antd';
import { type FC, type SetStateAction, useContext, useState } from 'react';
import { ErrorContext } from '../../../../errorHandling/ErrorContext';
import {
  EnvironmentType,
  Phase2,
  useDeleteInstanceMutation,
} from '../../../../generated-types';
import { type Instance, WorkspaceRole } from '../../../../utils';
import { ModalAlert } from '../../../common/ModalAlert';
import type { InstanceEnvironment } from '../../../../utils';
export interface IRowInstanceActionsDefaultProps {
  extended: boolean;
  instance: Instance;
  viewMode: WorkspaceRole;
  setSshModal: React.Dispatch<SetStateAction<boolean>>;
}

const RowInstanceActionsDefault: FC<IRowInstanceActionsDefaultProps> = ({
  ...props
}) => {
  const { extended, instance, viewMode, setSshModal } = props;
  const {
    prettyName,
    url,
    status,
    gui,
    name,
    tenantNamespace,
    environmentType,
    environments,
  } = instance;

  const { apolloErrorCatcher } = useContext(ErrorContext);

  const [deleteInstanceMutation] = useDeleteInstanceMutation({
    onError: apolloErrorCatcher,
  });

  const titleFromStatus = () => {
    if (!connectDisabled) {
      return gui ? 'Connect to the GUI' : `Connect through SSH`;
    } else
      return (
        <>
          <div>
            {status === Phase2.ResourceQuotaExceeded ? (
              <div>
                <b>You have reached your limit of resources</b>
                <br />
                Please delete an instance to create a new one
              </div>
            ) : (
              <div>
                <div>
                  <b>{'Impossible to connect:'}</b>
                </div>
                <div>
                  {environmentType === EnvironmentType.Container
                    ? 'Containers do not support SSH connection yet'
                    : `The instance is ${status}`}
                </div>
              </div>
            )}
          </div>
        </>
      );
  };

  const classFromProps = () => {
    if (!connectDisabled) {
      if (extended) return 'primary';
      else return 'green';
    }
    return 'primary';
  };

  const classFromPropsMobile = () => {
    if (!connectDisabled) {
      if (extended) return 'default';
      else return 'green';
    }
    return 'default';
  };

  const connectDisabled =
    status !== Phase2.Ready ||
    (environmentType === EnvironmentType.Container && !gui);

  const font22px = { fontSize: '22px' };

  const [showDeleteModalConfirm, setShowDeleteModalConfirm] = useState(false);

  const handleConnect = () => {
    if (environments && environments.length == 1) {
      const env = environments[0];
      handleEnvironmentConnect(env);
    }
  };

  const handleEnvironmentConnect = (env: InstanceEnvironment) => {
    if (env.guiEnabled) {
      const baseUrl = url?.endsWith('/') ? url.slice(0, -1) : url;
      const envUrl = `${baseUrl}/${env.name}/`;
      window.open(envUrl, '_blank');
    } else {
      setSshModal(true);
    }
  };

  // Dropdown menu items for environments
  const createEnvironmentMenuItems = (): MenuProps['items'] => {
    if (!environments || environments.length <= 1) return [];

    return environments.map(env => {
      const isReady = env.phase === Phase2.Ready;
      const isGuiEnabled = env.guiEnabled;

      return {
        key: env.name,
        label: env.name,
        icon: isGuiEnabled ? <ExportOutlined /> : <CodeOutlined />,
        disabled: !isReady,
        onClick: () => handleEnvironmentConnect(env),
      };
    });
  };

  const environmentMenuProps: MenuProps = {
    items: createEnvironmentMenuItems(),
  };

  return (
    <>
      <ModalAlert
        headTitle="Confirm Instance deletion"
        message={
          <>
            Do you really want to delete <b>{prettyName}</b> ?
          </>
        }
        description={`Instance ID: ${name}`}
        type="warning"
        buttons={[
          <Button
            key={0}
            shape="round"
            className="mr-2 w-24"
            type="primary"
            onClick={() => setShowDeleteModalConfirm(false)}
          >
            Close
          </Button>,
          <Button
            key={1}
            shape="round"
            className="ml-2 w-24"
            type="primary"
            danger
            onClick={() =>
              deleteInstanceMutation({
                variables: {
                  instanceId: name,
                  tenantNamespace: tenantNamespace!,
                },
              })
                .then(() => {
                  setShowDeleteModalConfirm(false);
                })
                .catch(() => null)
            }
          >
            Delete
          </Button>,
        ]}
        show={showDeleteModalConfirm}
        setShow={setShowDeleteModalConfirm}
      />
      <Tooltip placement="top" title="Destroy">
        <Button
          onClick={() => setShowDeleteModalConfirm(true)}
          type="link"
          danger
          className={`hidden ${
            extended ? 'sm:block' : 'xs:block'
          } py-0 border-0`}
          shape="circle"
          size="middle"
          icon={
            <DeleteOutlined
              className="flex justify-center items-center"
              style={font22px}
            />
          }
        />
      </Tooltip>
      <Tooltip placement="top" title={titleFromStatus()}>
        <div
          className={`hidden ${
            extended
              ? viewMode === WorkspaceRole.manager
                ? 'xl:block'
                : 'lg:block'
              : 'sm:block '
          } ${connectDisabled ? 'cursor-not-allowed' : ''}`}
        >
          {environments && environments.length > 1 ? (
            <Dropdown
              menu={environmentMenuProps}
              disabled={connectDisabled}
              trigger={['click']}
            >
              <Button
                type="primary"
                color={classFromProps()}
                variant="solid"
                shape="round"
                size="middle"
                disabled={connectDisabled}
                icon={<DownOutlined />}
              >
                Connect ({environments.length} envs)
              </Button>
            </Dropdown>
          ) : (
            <Button
              className={`${connectDisabled ? 'pointer-events-none' : ''}`}
              color={classFromProps()}
              type="primary"
              variant="solid"
              ghost={!gui}
              shape="round"
              size="middle"
              onClick={handleConnect}
              disabled={connectDisabled}
            >
              Connect
            </Button>
          )}
        </div>
        <div
          className={`hidden ${
            extended
              ? `sm:block ${
                  viewMode === WorkspaceRole.manager ? 'xl:hidden' : 'lg:hidden'
                }`
              : 'xs:block sm:hidden'
          } block flex items-center ${
            connectDisabled ? 'cursor-not-allowed' : ''
          }`}
        >
          {environments && environments.length > 1 ? (
            <Dropdown
              menu={environmentMenuProps}
              trigger={['click']}
              disabled={connectDisabled}
            >
              <Button
                className={`${
                  connectDisabled ? 'pointer-events-none' : ''
                } flex items-center justify-center p-0 border-0`}
                type={!extended ? 'link' : 'default'}
                color={classFromPropsMobile()}
                shape="circle"
                size="middle"
                disabled={connectDisabled}
                icon={
                  <DownOutlined
                    className="flex items-center justify-center"
                    style={font22px}
                  />
                }
              />
            </Dropdown>
          ) : (
            <Button
              className={`${
                connectDisabled ? 'pointer-events-none' : ''
              } flex items-center justify-center p-0 border-0`}
              type={!extended ? 'link' : 'default'}
              color={classFromPropsMobile()}
              shape="circle"
              size="middle"
              onClick={handleConnect}
              disabled={connectDisabled}
              icon={
                <ExportOutlined
                  className="flex items-center justify-center"
                  style={font22px}
                />
              }
            />
          )}
        </div>
      </Tooltip>
    </>
  );
};

export default RowInstanceActionsDefault;
