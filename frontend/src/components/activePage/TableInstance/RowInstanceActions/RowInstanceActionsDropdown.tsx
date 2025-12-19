import { type FC, type SetStateAction, useContext, useState } from 'react';
import { Dropdown, Badge, Space } from 'antd';
import { Button } from 'antd';
import { Link } from 'react-router-dom';
import {
  SelectOutlined,
  CodeOutlined,
  DeleteOutlined,
  FolderOpenOutlined,
  MoreOutlined,
  PoweroffOutlined,
  CaretRightOutlined,
  ExclamationCircleOutlined,
  ExportOutlined,
} from '@ant-design/icons';
import type { Instance } from '../../../../utils';
import {
  EnvironmentType,
  Phase2,
  useApplyInstanceMutation,
  useDeleteInstanceMutation,
} from '../../../../generated-types';
import { setInstanceRunning } from '../../../../utilsLogic';
import { ErrorContext } from '../../../../errorHandling/ErrorContext';

export interface IRowInstanceActionsDropdownProps {
  instance: Instance;
  fileManager?: boolean;
  extended: boolean;
  setSshModal: React.Dispatch<SetStateAction<boolean>>;
  onEnablePublicExposure?: () => void;
}

const RowInstanceActionsDropdown: FC<IRowInstanceActionsDropdownProps> = ({
  ...props
}) => {
  const {
    instance,
    fileManager,
    extended,
    setSshModal,
    onEnablePublicExposure,
  } = props;

  const {
    status,
    persistent,
    url,
    name,
    tenantNamespace,
    environmentType,
    gui,
  } = instance;

  const font20px = { fontSize: '20px' };

  const [disabled, setDisabled] = useState(false);
  const { apolloErrorCatcher } = useContext(ErrorContext);
  const [deleteInstanceMutation] = useDeleteInstanceMutation({
    onError: apolloErrorCatcher,
  });
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
      } catch {
        // TODO: do nothing at the moment
      }
    }
  };

  const statusComponents = {
    [Phase2.Ready]: {
      menuIcon: <PoweroffOutlined style={font20px} />,
      menuText: 'Stop',
      menuAction: () => mutateInstanceStatus(false),
    },
    [Phase2.Off]: {
      menuIcon: <CaretRightOutlined style={font20px} />,
      menuText: 'Start',
      menuAction: () => mutateInstanceStatus(true),
    },
    Other: {
      menuIcon: <ExclamationCircleOutlined style={font20px} />,
      menuText: '',
      menuAction: () => null,
    },
  } as const;

  const { menuIcon, menuText, menuAction } =
    status === Phase2.Ready
      ? statusComponents[Phase2.Ready]
      : status === Phase2.Off
        ? statusComponents[Phase2.Off]
        : statusComponents.Other;

  const isContainer =
    environmentType === EnvironmentType.Container ||
    environmentType === EnvironmentType.Standalone;

  const sshDisabled = status !== Phase2.Ready || isContainer;

  const fileManagerDisabled = status !== Phase2.Ready && isContainer;

  const connectDisabled = status !== Phase2.Ready || (isContainer && !gui);

  const getFirstEnvironmentName = () => {
    return instance.environments?.[0]?.name || 'env';
  };

  const buildSSHLink = (envName: string) => {
    if (envName) {
      return `/instance/${instance.tenantNamespace}/${instance.name}/${envName}/ssh`;
    }
    return `/instance/${instance.tenantNamespace}/${instance.name}/env/ssh`;
  };

  return (
    <Dropdown
      trigger={['click']}
      menu={{
        items: [
          {
            key: 'connect',
            label: 'Connect',
            disabled: connectDisabled,
            icon: <SelectOutlined style={font20px} />,
            onClick: gui
              ? () => window.open(url!, '_blank')
              : () => setSshModal(true),
            className: `flex items-center sm:hidden ${
              !connectDisabled
                ? extended
                  ? 'primary-color-fg'
                  : 'success-color-fg xs:hidden'
                : 'pointer-events-none'
            }`,
          },
          persistent
            ? {
                key: 'persistent',
                label: menuText,
                icon: menuIcon,
                onClick: () => menuAction,
                className: `flex items-center ${
                  extended ? ' sm:hidden' : 'xs:hidden'
                }`,
              }
            : null,
          {
            type: 'divider',
            className: `${extended ? 'sm:hidden' : 'xs:hidden'}`,
          },
          ...(onEnablePublicExposure
            ? [
                {
                  key: 'expose',
                  label: (
                    <Space align="center">
                      Port Exposure
                      {instance.publicExposure &&
                        (instance.publicExposure?.ports ?? []).length > 0 && (
                          <Badge
                            count={
                              (instance.publicExposure?.ports ?? []).length
                            }
                            showZero={false}
                            size="small"
                          />
                        )}
                    </Space>
                  ),
                  icon: <SelectOutlined style={font20px} />,
                  onClick: () => onEnablePublicExposure?.(),
                },
                {
                  type: 'divider' as const,
                },
              ]
            : []),
          {
            key: 'ssh',
            icon: <CodeOutlined style={font20px} />,
            disabled: sshDisabled,
            label: (
              <>
                SSH
                {/* Only show direct link button if there's exactly one environment */}
                {(!instance.environments ||
                  instance.environments.length === 1) && (
                  <Button
                    disabled={sshDisabled}
                    type="link"
                    className="ml-3"
                    color="primary"
                    variant="solid"
                    shape="circle"
                    size="small"
                    icon={
                      <Link
                        to={buildSSHLink(getFirstEnvironmentName())}
                        target="_blank"
                        rel="noopener noreferrer"
                        onClick={e => e.stopPropagation()}
                        style={{
                          color: 'inherit',
                          display: 'flex',
                          alignItems: 'center',
                        }}
                      >
                        <span style={{ filter: 'drop-shadow(0 0 0 black)' }}>
                          <ExportOutlined style={{ fontSize: 15 }} />
                        </span>
                      </Link>
                    }
                  ></Button>
                )}
              </>
            ),
            onClick: () => setSshModal(true),
            className: `flex items-center ${extended ? 'xl:hidden' : ''} ${sshDisabled ? 'pointer-events-none' : ''}`,
          },
          {
            key: 'upload',
            label: isContainer
              ? 'File Manager'
              : environmentType === EnvironmentType.VirtualMachine
                ? 'Drive'
                : '',
            icon: <FolderOpenOutlined style={font20px} />,
            disabled: fileManagerDisabled,
            className: `flex items-center ${extended ? 'xl:hidden' : ''} `,
            onClick: () => {},
          },
          {
            type: 'divider',
            className: `${extended ? 'sm:hidden' : 'xs:hidden'}`,
          },
          {
            key: 'destroy',
            label: 'Destroy',
            danger: true,
            icon: <DeleteOutlined style={font20px} />,
            onClick: () =>
              deleteInstanceMutation({
                variables: {
                  instanceId: name,
                  tenantNamespace: tenantNamespace!,
                },
              }),
            className: `flex items-center ${
              extended ? ' sm:hidden' : 'xs:hidden'
            }`,
          },
        ],
      }}
    >
      <Button
        className={`${
          extended
            ? !sshDisabled || fileManager
              ? 'xl:hidden'
              : 'sm:hidden'
            : ''
        } flex justify-center items-center`}
        color="default"
        type="link"
        shape="circle"
        size="middle"
        icon={<MoreOutlined className="flex items-center" style={font20px} />}
      />
    </Dropdown>
  );
};

export default RowInstanceActionsDropdown;
