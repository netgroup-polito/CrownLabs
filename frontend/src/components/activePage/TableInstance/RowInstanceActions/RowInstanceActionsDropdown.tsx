import { FC, SetStateAction, useContext, useState } from 'react';
import { Dropdown, Menu } from 'antd';
import Button from 'antd-button-color';
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
import { Instance } from '../../../../utils';
import {
  EnvironmentType,
  Phase,
  useApplyInstanceMutation,
  useDeleteInstanceMutation,
} from '../../../../generated-types';
import { DropDownAction, setInstanceRunning } from '../../../../utilsLogic';
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
    myDriveUrl,
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
          applyInstanceMutation
        );
        if (result) setTimeout(setDisabled, 400, false);
      } catch {
        // TODO: do nothing at the moment
      }
    }
  };

  const statusComponents = {
    Ready: {
      menuKey: 'stop',
      menuIcon: <PoweroffOutlined style={font20px} />,
      menuText: 'Stop',
    },
    Off: {
      menuKey: 'start',
      menuIcon: <CaretRightOutlined style={font20px} />,
      menuText: 'Start',
    },
    Other: {
      menuKey: '',
      menuIcon: <ExclamationCircleOutlined style={font20px} />,
      menuText: '',
    },
  };

  const { menuKey, menuIcon, menuText } =
    statusComponents[
      status === Phase.Ready || status === Phase.Off ? status : 'Other'
    ];

  const dropdownHandler = (key: DropDownAction) => {
    switch (key) {
      case DropDownAction.start:
        persistent && mutateInstanceStatus(true);
        break;
      case DropDownAction.stop:
        persistent && mutateInstanceStatus(false);
        break;
      case DropDownAction.connect:
        gui ? window.open(url!, '_blank') : setSshModal(true);
        break;
      case DropDownAction.destroy:
        deleteInstanceMutation({
          variables: {
            instanceId: name,
            tenantNamespace: tenantNamespace!,
          },
        });
        break;
      case DropDownAction.ssh:
        setSshModal(true);
        break;
      case DropDownAction.upload:
        (environmentType === EnvironmentType.Container ||
          environmentType === EnvironmentType.Standalone) &&
          window.open(`${myDriveUrl}/files`, '_blank');
        environmentType === EnvironmentType.VirtualMachine &&
          window.open('https://crownlabs.polito.it/cloud', '_blank');
        break;
      default:
        break;
    }
  };

  const sshDisabled =
    status !== Phase.Ready ||
    environmentType === EnvironmentType.Container ||
    environmentType === EnvironmentType.Standalone;

  const fileManagerDisabled =
    status !== Phase.Ready &&
    (environmentType === EnvironmentType.Container ||
      environmentType === EnvironmentType.Standalone);

  const connectDisabled =
    status !== Phase.Ready ||
    ((environmentType === EnvironmentType.Container ||
      environmentType === EnvironmentType.Standalone) &&
      !gui);

  return (
    <Dropdown
      trigger={['click']}
      overlay={
        <Menu onClick={({ key }) => dropdownHandler(key as DropDownAction)}>
          <Menu.Item
            disabled={connectDisabled}
            key="connect"
            className={`flex items-center sm:hidden ${
              !connectDisabled
                ? extended
                  ? 'primary-color-fg'
                  : 'success-color-fg xs:hidden'
                : 'pointer-events-none'
            }`}
            icon={<ExportOutlined style={font20px} />}
          >
            Connect
          </Menu.Item>
          {persistent && (
            <Menu.Item
              key={menuKey}
              className={`flex items-center ${
                extended ? ' sm:hidden' : 'xs:hidden'
              }`}
              icon={menuIcon}
            >
              {menuText}
            </Menu.Item>
          )}
          <Menu.Divider className={`${extended ? 'sm:hidden' : 'xs:hidden'}`} />
          <Menu.Item
            key="ssh"
            className={`flex items-center ${extended ? 'xl:hidden' : ''} `}
            disabled={sshDisabled}
            icon={<CodeOutlined style={font20px} />}
          >
            SSH
          </Menu.Item>
          <Menu.Item
            key="upload"
            className={`flex items-center ${extended ? 'xl:hidden' : ''} `}
            disabled={fileManagerDisabled}
            icon={<FolderOpenOutlined style={font20px} />}
          >
            {environmentType === EnvironmentType.Container ||
            environmentType === EnvironmentType.Standalone
              ? 'File Manager'
              : environmentType === EnvironmentType.VirtualMachine && 'Drive'}
          </Menu.Item>
          <Menu.Divider className={`${extended ? 'sm:hidden' : 'xs:hidden'}`} />
          <Menu.Item
            key="destroy"
            className={`flex items-center ${
              extended ? ' sm:hidden' : 'xs:hidden'
            }`}
            danger
            icon={<DeleteOutlined style={font20px} />}
          >
            Destroy
          </Menu.Item>
        </Menu>
      }
    >
      <Button
        className={`${
          extended
            ? !sshDisabled || fileManager
              ? 'xl:hidden'
              : 'sm:hidden'
            : ''
        } flex justify-center items-center`}
        type="default"
        with="link"
        shape="circle"
        size="middle"
        icon={<MoreOutlined className="flex items-center" style={font20px} />}
      />
    </Dropdown>
  );
};

export default RowInstanceActionsDropdown;
