import { FC, Dispatch, SetStateAction } from 'react';
import { Upload, Dropdown, Menu, message } from 'antd';
import Button from 'antd-button-color';
import {
  CodeOutlined,
  DeleteOutlined,
  FolderOpenOutlined,
  MoreOutlined,
} from '@ant-design/icons';
import { ISSHInfo } from '../InstancesTable/InstancesTable';

export interface IInstanceActionsProps {
  phase: 'ready' | 'creating' | 'failed' | 'stopping' | 'off';
  toggleModal: Dispatch<SetStateAction<boolean>>;
  setSshInfo: Dispatch<SetStateAction<ISSHInfo>>;
  ip: string;
}

const InstanceActions: FC<IInstanceActionsProps> = ({ ...props }) => {
  const { phase, setSshInfo, toggleModal, ip } = props;
  return (
    <div className="flex justify-end items-center gap-2 pr-2">
      <Button
        shape="round"
        className="hidden lg:inline-block"
        disabled={phase !== 'ready'}
        onClick={() => {
          toggleModal(true);
          setSshInfo(old => ({ ...old, ip }));
        }}
      >
        SSH
      </Button>
      {/* <Upload name="file">
        <Button
          shape="circle"
          className="hidden lg:inline-block"
          disabled={phase !== 'ready'}
        >
          <FolderOpenOutlined />
        </Button>
      </Upload> */}
      <Button type="primary" shape="round" disabled={phase !== 'ready'}>
        Connect
      </Button>
      <Dropdown
        placement="bottomCenter"
        trigger={['click']}
        disabled={phase !== 'ready'}
        overlay={
          <Menu>
            <Menu.Item danger onClick={() => message.info('VM deleted')}>
              Confirm
            </Menu.Item>
          </Menu>
        }
      >
        <Button type="danger" shape="circle">
          <DeleteOutlined />
        </Button>
      </Dropdown>
      {phase === 'ready' && (
        <Dropdown
          trigger={['click']}
          overlay={
            <Menu>
              <Upload>
                <Menu.Item
                  icon={<FolderOpenOutlined style={{ fontSize: '18px' }} />}
                >
                  Upload
                </Menu.Item>
              </Upload>
              <Menu.Item icon={<CodeOutlined style={{ fontSize: '18px' }} />}>
                SSH
              </Menu.Item>
            </Menu>
          }
        >
          <MoreOutlined className="lg:hidden" />
        </Dropdown>
      )}
    </div>
  );
};

export default InstanceActions;
