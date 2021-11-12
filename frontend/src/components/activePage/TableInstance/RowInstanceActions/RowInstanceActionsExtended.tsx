import { FC, SetStateAction } from 'react';
import { Popover, Tooltip, Upload, Typography } from 'antd';
import Button from 'antd-button-color';
import { InfoOutlined, FolderOpenOutlined } from '@ant-design/icons';
import { VmStatus } from '../../../../utils';
import { EnvironmentType } from '../../../../generated-types';

const { Text } = Typography;

const getSSHTooltipText = (
  isInstanceReady: boolean,
  environmentType: EnvironmentType
) => {
  if (environmentType === EnvironmentType.Container)
    return 'Containers does not support SSH connection (yet!)';
  if (!isInstanceReady)
    return 'Instance must be ready in order to connect through SSH';
  return 'Show SSH connection instructions';
};

export interface IRowInstanceActionsExtendedProps {
  ip: string;
  time: string;
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
    templateName,
    environmentType,
    status,
    fileManager,
    setSshModal,
  } = props;

  const sshDisabled =
    status !== 'VmiReady' || environmentType === EnvironmentType.Container;

  const infoContent = (
    <>
      <p className="m-0">
        <strong>IP:</strong> <Text copyable>{ip}</Text>
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

        <Tooltip
          title={getSSHTooltipText(
            status === 'VmiReady',
            environmentType!
            //isOwnedInstance
          )}
        >
          <span className={`${sshDisabled ? 'cursor-not-allowed' : ''}`}>
            <Button
              disabled={sshDisabled}
              className={`hidden mr-3 xl:inline-block ${
                sshDisabled ? 'pointer-events-none' : ''
              }`}
              shape="round"
              onClick={() => setSshModal(true)}
            >
              SSH
            </Button>
          </span>
        </Tooltip>

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
