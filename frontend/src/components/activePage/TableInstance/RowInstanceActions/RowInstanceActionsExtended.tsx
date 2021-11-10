import { FC, SetStateAction } from 'react';
import { Popover, Tooltip, Upload } from 'antd';
import Button from 'antd-button-color';
import { InfoOutlined, FolderOpenOutlined } from '@ant-design/icons';
import { VmStatus } from '../../../../utils';
import { EnvironmentType } from '../../../../generated-types';

const getSSHTooltipText = (
  hasSSHKeys: boolean,
  isInstanceReady: boolean,
  environmentType: EnvironmentType
) => {
  if (environmentType === EnvironmentType.Container)
    return 'Containers does not support SSH connection (yet!)';
  if (!hasSSHKeys)
    return 'You have no SSH keys associated with your account, go to Account page and add them';
  if (!isInstanceReady)
    return 'Instance must be ready in order to connect through SSH';
  return 'Show SSH connection instructions';
};

export interface IRowInstanceActionsExtendedProps {
  ip: string;
  time: string;
  hasSSHKeys?: boolean;
  templateName: string;
  environmentType?: EnvironmentType;
  status: VmStatus;
  fileManager?: boolean;
  setSshModal: React.Dispatch<SetStateAction<boolean>>;
}

const RowInstanceActionsExtended: FC<IRowInstanceActionsExtendedProps> = ({
  ...props
}) => {
  const {
    ip,
    time,
    hasSSHKeys,
    templateName,
    environmentType,
    status,
    fileManager,
    setSshModal,
  } = props;

  const infoContent = (
    <>
      <p className="m-0">
        <strong>IP:</strong> {ip}
      </p>
      <p className="m-0 lg:hidden">
        <strong>Created:</strong> {time} ago
      </p>
      <p className="m-0 md:hidden">
        <strong>Template:</strong> {templateName}
      </p>
    </>
  );
  return (
    <>
      <div className="inline-flex border-box justify-center">
        <Popover placement="top" content={infoContent} trigger="click">
          <Button
            shape="circle"
            className="hidden sm:block mr-3"
            disabled={status !== 'VmiReady'}
          >
            <InfoOutlined />
          </Button>
        </Popover>
        {!hasSSHKeys ||
        status !== 'VmiReady' ||
        environmentType === EnvironmentType.Container ? (
          <Tooltip
            title={getSSHTooltipText(
              hasSSHKeys || false,
              status === 'VmiReady',
              environmentType!
            )}
          >
            <span className="cursor-not-allowed">
              <Button
                disabled
                className="hidden xl:inline-block mr-3 pointer-events-none"
                shape="round"
              >
                Connect via SSH
              </Button>
            </span>
          </Tooltip>
        ) : (
          <Button
            shape="round"
            className="hidden xl:inline-block mr-3"
            onClick={() => setSshModal(true)}
          >
            Connect via SSH
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
                  className={`hidden mr-3 xl:inline-block ${
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
      </div>
    </>
  );
};

export default RowInstanceActionsExtended;
