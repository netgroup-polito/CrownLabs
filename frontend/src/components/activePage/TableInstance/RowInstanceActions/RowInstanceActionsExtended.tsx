import { FC, SetStateAction } from 'react';
import { Popover, Tooltip, Space, Upload } from 'antd';
import Button from 'antd-button-color';
import { InfoOutlined, FolderOpenOutlined } from '@ant-design/icons';
import { VmStatus } from '../../../../utils';
import { ISSHInfo } from '../SSHModalContent/SSHModalContent';

export interface IRowInstanceActionsExtendedProps {
  ip: string;
  time: string;
  ssh?: ISSHInfo;
  status: VmStatus;
  fileManager?: boolean;
  setSshModal: React.Dispatch<SetStateAction<boolean>>;
}

const RowInstanceActionsExtended: FC<IRowInstanceActionsExtendedProps> = ({
  ...props
}) => {
  const { ip, time, ssh, status, fileManager, setSshModal } = props;

  const infoContent = (
    <>
      <p className="m-0">
        <strong>IP:</strong> {ip}
      </p>
      <p className="m-0 lg:hidden">
        <strong>Created by:</strong> {time}
      </p>
    </>
  );
  return (
    <>
      <Space size={'middle'}>
        <Popover placement="top" content={infoContent} trigger="click">
          <Button
            shape="circle"
            className="hidden md:block "
            disabled={status !== 'VmiReady'}
          >
            <InfoOutlined />
          </Button>
        </Popover>
        {ssh && (
          <Button
            shape="round"
            className="hidden xl:inline-block"
            disabled={status !== 'VmiReady'}
            onClick={() => setSshModal(true)}
          >
            SSH
          </Button>
        )}

        {fileManager && (
          <Tooltip placement="top" title={'File Manager'}>
            <Upload name="file">
              <span
                className={`${
                  status !== 'VmiReady' ? 'cursor-not-allowed' : ''
                }`}
              >
                <Button
                  shape="circle"
                  className={`hidden xl:inline-block ${
                    status !== 'VmiReady' ? 'pointer-events-none' : ''
                  }`}
                  disabled={status !== 'VmiReady'}
                >
                  <FolderOpenOutlined />
                </Button>
              </span>
            </Upload>
          </Tooltip>
        )}
      </Space>
    </>
  );
};

export default RowInstanceActionsExtended;
