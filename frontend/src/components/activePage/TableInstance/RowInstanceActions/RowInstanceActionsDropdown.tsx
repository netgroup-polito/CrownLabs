import { FC, SetStateAction } from 'react';
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
} from '@ant-design/icons';
import { ISSHInfo } from '../SSHModalContent/SSHModalContent';
import { Instance } from '../../../../utils';

export interface IRowInstanceActionsDropdownProps {
  instance: Instance;
  fileManager?: boolean;
  ssh?: ISSHInfo;
  extended: boolean;
  startInstance?: (idInstance: string, idTemplate: string) => void;
  stopInstance?: (idInstance: string, idTemplate: string) => void;
  destroyInstance: (instanceId: string, tenantNamespace: string) => void;
  setSshModal: React.Dispatch<SetStateAction<boolean>>;
}

const RowInstanceActionsDropdown: FC<IRowInstanceActionsDropdownProps> = ({
  ...props
}) => {
  const {
    instance,
    fileManager,
    ssh,
    extended,
    startInstance,
    stopInstance,
    destroyInstance,
    setSshModal,
  } = props;

  const {
    status,
    persistent,
    url,
    idTemplate,
    name,
    tenantDisplayName,
  } = instance;

  const dropdownHandler = (key: string) => {
    switch (key) {
      case 'Start':
        persistent && startInstance?.(name, idTemplate!);
        break;
      case 'Stop':
        persistent && stopInstance?.(name, idTemplate!);
        break;
      case 'Connect':
        window.open(url!, '_blank');
        break;
      case 'Destroy':
        destroyInstance(name!, tenantDisplayName!);
        break;
      case 'SSH':
        setSshModal(true);
        break;
      case 'Upload':
        // Something to add
        break;
      default:
        break;
    }
  };

  const font20px = { fontSize: '20px' };

  return (
    <Dropdown
      trigger={['click']}
      overlay={
        <Menu onClick={click => dropdownHandler(click.key.toString())}>
          <Menu.Item
            disabled={status !== 'VmiReady'}
            key="Connect"
            className={`sm:hidden ${
              status === 'VmiReady' &&
              (extended ? 'primary-color-fg' : 'success-color-fg')
            }`}
            icon={<ExportOutlined style={font20px} />}
          >
            Connect
          </Menu.Item>
          {persistent && startInstance && status === 'VmiOff' && (
            <Menu.Item
              key="Start"
              className="sm:hidden"
              icon={<CaretRightOutlined style={font20px} />}
            >
              Start
            </Menu.Item>
          )}
          {persistent && stopInstance && status === 'VmiReady' && (
            <Menu.Item
              key="Stop"
              className="sm:hidden"
              icon={<PoweroffOutlined style={font20px} />}
            >
              Stop
            </Menu.Item>
          )}
          {extended && (ssh || fileManager) && (
            <Menu.Divider className={`${extended ? 'sm:hidden' : 'hidden'}`} />
          )}
          {extended && fileManager && (
            <Menu.Item
              key="Upload"
              className="xl:hidden"
              disabled={status !== 'VmiReady'}
              icon={<FolderOpenOutlined style={font20px} />}
            >
              File Manager
            </Menu.Item>
          )}
          {extended && ssh && (
            <Menu.Item
              key="SSH"
              className="xl:hidden"
              disabled={status !== 'VmiReady'}
              icon={<CodeOutlined style={font20px} />}
            >
              SSH
            </Menu.Item>
          )}
          <Menu.Divider className={`${extended ? 'sm:hidden' : 'xs:hidden'}`} />
          <Menu.Item
            key="Destroy"
            className="sm:hidden"
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
