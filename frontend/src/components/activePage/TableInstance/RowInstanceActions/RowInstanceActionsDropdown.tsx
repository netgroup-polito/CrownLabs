import { FC, SetStateAction, useState } from 'react';
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
import { ISSHInfo } from '../SSHModalContent/SSHModalContent';
import { Instance } from '../../../../utils';
import {
  useApplyInstanceMutation,
  useDeleteInstanceMutation,
} from '../../../../generated-types';
import { DropDownAction, setInstanceRunning } from '../../ActiveUtils';

export interface IRowInstanceActionsDropdownProps {
  instance: Instance;
  fileManager?: boolean;
  ssh?: ISSHInfo;
  extended: boolean;
  setSshModal: React.Dispatch<SetStateAction<boolean>>;
}

const RowInstanceActionsDropdown: FC<IRowInstanceActionsDropdownProps> = ({
  ...props
}) => {
  const { instance, fileManager, ssh, extended, setSshModal } = props;

  const { status, persistent, url, name, tenantNamespace } = instance;

  const font20px = { fontSize: '20px' };

  const [disabled, setDisabled] = useState(false);
  const [deleteInstanceMutation] = useDeleteInstanceMutation();
  const [applyInstanceMutation] = useApplyInstanceMutation();

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
    VmiReady: {
      menuKey: 'stop',
      menuIcon: <PoweroffOutlined style={font20px} />,
      menuText: 'Stop',
    },
    VmiOff: {
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

  const { menuKey, menuIcon, menuText } = statusComponents[
    status === 'VmiReady' || status === 'VmiOff' ? status : 'Other'
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
        window.open(url!, '_blank');
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
        // TODO: Something to add
        break;
      case DropDownAction.destroy_all:
        // TODO: Popconfirm not work maybe we should use a modal for the confirmation
        break;
      default:
        break;
    }
  };

  return (
    <Dropdown
      trigger={['click']}
      overlay={
        <Menu onClick={({ key }) => dropdownHandler(key as DropDownAction)}>
          <Menu.Item
            disabled={status !== 'VmiReady'}
            key="connect"
            className={`flex items-center sm:hidden ${
              status === 'VmiReady' &&
              (extended ? 'primary-color-fg' : 'success-color-fg')
            }`}
            icon={<ExportOutlined style={font20px} />}
          >
            Connect
          </Menu.Item>
          {persistent && (
            <Menu.Item
              key={menuKey}
              className="flex items-center sm:hidden"
              icon={menuIcon}
            >
              {menuText}
            </Menu.Item>
          )}
          {extended && (ssh || fileManager) && (
            <Menu.Divider className={`${extended ? 'sm:hidden' : 'hidden'}`} />
          )}
          {extended && fileManager && (
            <Menu.Item
              key="upload"
              className="flex items-center xl:hidden"
              disabled={status !== 'VmiReady'}
              icon={<FolderOpenOutlined style={font20px} />}
            >
              File Manager
            </Menu.Item>
          )}
          {extended && ssh && (
            <Menu.Item
              key="ssh"
              className="flex items-center xl:hidden"
              disabled={status !== 'VmiReady'}
              icon={<CodeOutlined style={font20px} />}
            >
              SSH
            </Menu.Item>
          )}
          <Menu.Divider className={`${extended ? 'sm:hidden' : 'xs:hidden'}`} />
          <Menu.Item
            key="destroy"
            className="flex items-center sm:hidden"
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
            ? ssh || fileManager
              ? 'xl:hidden'
              : 'sm:hidden'
            : 'xs:hidden'
        } flex justify-center`}
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
