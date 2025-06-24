import type { FC, SetStateAction } from 'react';
import { Popover, Tooltip, Typography } from 'antd';
import { Button } from 'antd';
import { InfoOutlined } from '@ant-design/icons';
import { type Instance, WorkspaceRole } from '../../../../utils';
import { EnvironmentType, Phase } from '../../../../generated-types';
import { Link } from 'react-router-dom';
import { ExportOutlined } from '@ant-design/icons';
const { Text } = Typography;

const getSSHTooltipText = (
  isInstanceReady: boolean,
  environmentType: EnvironmentType,
) => {
  if (environmentType === EnvironmentType.Standalone)
    return 'Standalone applications do not support SSH connection (yet!)';
  if (environmentType === EnvironmentType.Container)
    return 'Containers do not support SSH connection (yet!)';
  if (!isInstanceReady)
    return 'Instance must be ready in order to connect through SSH';
  return 'Show SSH connection instructions';
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
    name,
    prettyName,
    nodeName,
    running,
  } = instance;

  const sshDisabled =
    status !== Phase.Ready ||
    environmentType === EnvironmentType.Container ||
    environmentType === EnvironmentType.Standalone;

  const ENV_PLACEHOLDER = 'env';

  const infoContent = (
    <>
      <p className="m-0">
        <strong>Instance ID: </strong>
        <Text italic>{name}</Text>
      </p>
      {running && (
        <>
          <p className="m-0">
            <strong>IP: </strong>
            <Text type="warning" copyable={!!ip}>
              {ip ?? 'unknown'}
            </Text>
          </p>
          <p className="m-0">
            <strong>Node: </strong>
            <Text type="warning">{nodeName ?? '[choosing...]'}</Text>
          </p>
        </>
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
      <div className="inline-flex border-box justify-center xl:pl-4">
        <Popover placement="top" content={infoContent} trigger="click">
          <Button shape="circle" className="hidden sm:block mr-3">
            <InfoOutlined />
          </Button>
        </Popover>

        <Tooltip
          title={getSSHTooltipText(status === Phase.Ready, environmentType!)}
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
                    to={`/instance/${instance.tenantNamespace}/${instance.name}/${ENV_PLACEHOLDER}/ssh`}
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
            </Button>
          </span>
        </Tooltip>
      </div>
    </>
  );
};

export default RowInstanceActionsExtended;
