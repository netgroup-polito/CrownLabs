import { FC, SetStateAction } from 'react';
import { Popover, Tooltip, Typography } from 'antd';
import Button from 'antd-button-color';
import { InfoOutlined, FolderOpenOutlined } from '@ant-design/icons';
import { Instance, WorkspaceRole } from '../../../../utils';
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

const getFileManagerTooltipText = (
  isInstanceReady: boolean,
  environmentType: EnvironmentType
) => {
  if (environmentType === EnvironmentType.VirtualMachine) {
    return 'Drive';
  }
  if (environmentType === EnvironmentType.Container) {
    if (!isInstanceReady)
      return 'Instance must be ready in order to manage this Container files';
    else return 'File Manager';
  }
};

export interface IRowInstanceActionsExtendedProps {
  instance: Instance;
  time: string;
  viewMode: WorkspaceRole;
  setSshModal: React.Dispatch<SetStateAction<boolean>>;
}

const RowInstanceActionsExtended: FC<IRowInstanceActionsExtendedProps> = ({
  ...props
}) => {
  const { instance, time, viewMode, setSshModal } = props;
  const {
    ip,
    environmentType,
    status,
    templatePrettyName,
    url,
    name,
    prettyName,
  } = instance;

  const sshDisabled =
    status !== 'VmiReady' || environmentType === EnvironmentType.Container;

  const fileManagerDisabled =
    status !== 'VmiReady' && environmentType === EnvironmentType.Container;

  const infoContent = (
    <>
      <p className="m-0">
        <strong>IP: </strong>
        <Text type="warning" copyable={!!ip}>
          {ip ?? 'unknown'}
        </Text>
      </p>
      {viewMode === WorkspaceRole.manager && (
        <p className="m-0">
          <strong>Instance ID: </strong>
          <Text italic>{name}</Text>
        </p>
      )}
      {viewMode === WorkspaceRole.manager && (
        <p className="m-0 lg:hidden">
          <strong>PrettyName: </strong>
          <Text italic>{prettyName ?? 'unknown'}</Text>
        </p>
      )}
      <p className="m-0 lg:hidden">
        <strong>Created: </strong> {time ?? 'unknown'} <Text italic>ago</Text>
      </p>
      {viewMode !== WorkspaceRole.manager && (
        <p className="m-0 md:hidden">
          <strong>Template: </strong>
          <Text italic>{templatePrettyName ?? 'unknown'}</Text>
        </p>
      )}
    </>
  );
  return (
    <>
      <div className="inline-flex border-box justify-center">
        <Popover placement="top" content={infoContent} trigger="click">
          <Button shape="circle" className="hidden sm:block mr-3">
            <InfoOutlined />
          </Button>
        </Popover>

        <Tooltip
          title={getSSHTooltipText(status === 'VmiReady', environmentType!)}
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
        <Tooltip
          title={getFileManagerTooltipText(
            status === 'VmiReady',
            environmentType!
          )}
        >
          <span
            className={`${fileManagerDisabled ? 'cursor-not-allowed' : ''}`}
          >
            <Button
              shape="circle"
              className={`hidden mr-3 xl:inline-block ${
                fileManagerDisabled ? 'pointer-events-none' : ''
              }`}
              disabled={fileManagerDisabled}
              href={
                environmentType === EnvironmentType.Container
                  ? `${url}/mydrive/files`
                  : 'https://crownlabs.polito.it/cloud'
              }
              target="_blank"
            >
              <FolderOpenOutlined />
            </Button>
          </span>
        </Tooltip>
      </div>
    </>
  );
};

export default RowInstanceActionsExtended;
