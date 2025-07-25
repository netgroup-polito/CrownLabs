import { type FC, type SetStateAction, useContext, useState } from 'react';
import { Dropdown } from 'antd';
import { Button } from 'antd';
import {
  ExportOutlined,
  CodeOutlined,
  DeleteOutlined,
  FolderOpenOutlined,
  MoreOutlined,
  PoweroffOutlined,
  CaretRightOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';
import type { Instance } from '../../../../utils';
import {
  EnvironmentType,
  Phase,
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
}

const RowInstanceActionsDropdown: FC<IRowInstanceActionsDropdownProps> = ({
  ...props
}) => {
  const { instance, fileManager, extended, setSshModal } = props;

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
    [Phase.Ready]: {
      menuIcon: <PoweroffOutlined style={font20px} />,
      menuText: 'Stop',
      menuAction: () => mutateInstanceStatus(false),
    },
    [Phase.Off]: {
      menuIcon: <CaretRightOutlined style={font20px} />,
      menuText: 'Start',
      menuAction: () => mutateInstanceStatus(true),
    },
    Other: {
      menuIcon: <ExclamationCircleOutlined style={font20px} />,
      menuText: '',
      menuAction: () => null,
    },
  };

  const { menuIcon, menuText, menuAction } =
    status === Phase.Ready || status === Phase.Off
      ? statusComponents[status]
      : statusComponents.Other;

  const isContainer =
    environmentType === EnvironmentType.Container ||
    environmentType === EnvironmentType.Standalone;

  const sshDisabled = status !== Phase.Ready || isContainer;

  const fileManagerDisabled = status !== Phase.Ready && isContainer;

  const connectDisabled = status !== Phase.Ready || (isContainer && !gui);

  return (
    <Dropdown
      trigger={['click']}
      menu={{
        items: [
          {
            key: 'connect',
            label: 'Connect',
            disabled: connectDisabled,
            icon: <ExportOutlined style={font20px} />,
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
          {
            key: 'ssh',
            label: 'SSH',
            icon: <CodeOutlined style={font20px} />,
            onClick: () => setSshModal(true),
            className: `flex items-center ${
              extended ? 'xl:hidden' : ''
            } ${sshDisabled ? 'pointer-events-none' : ''}`,
            disabled: sshDisabled,
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
