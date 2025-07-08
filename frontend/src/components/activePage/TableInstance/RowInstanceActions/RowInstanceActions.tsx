import { type FC, useState } from 'react';
import { Modal, Typography } from 'antd';
import { Button } from 'antd';
import { type Instance, WorkspaceRole } from '../../../../utils';
import RowInstanceActionsPersistent from './RowInstanceActionsPersistent';
import RowInstanceActionsDropdown from './RowInstanceActionsDropdown';
import RowInstanceActionsExtended from './RowInstanceActionsExtended';
import SSHModalContent from '../SSHModalContent/SSHModalContent';
import RowInstanceActionsDefault from './RowInstanceActionsDefault';

const { Text } = Typography;
export interface IRowInstanceActionsProps {
  instance: Instance;
  now: Date;
  fileManager?: boolean;
  hasSSHKeys?: boolean;
  extended: boolean;
  viewMode: WorkspaceRole;
}

const RowInstanceActions: FC<IRowInstanceActionsProps> = ({ ...props }) => {
  const { instance, now, fileManager, hasSSHKeys, extended, viewMode } = props;

  const { persistent } = instance;

  const [sshModal, setSshModal] = useState(false);

  const getTime = () => {
    if (!instance.timeStamp) return 'unknown';
    const timeStamp = new Date(instance.timeStamp!);
    // Get Delta time
    let delta = (now.getTime() - timeStamp.getTime()) / 1000;
    // Get Years
    const years = Math.floor(delta / (86400 * 365));
    // Get Days
    delta -= years * (86400 * 365);
    const days = Math.floor(delta / 86400);
    // Get hours
    delta -= days * 86400;
    const hours = Math.floor(delta / 3600) % 24;
    // Get Minutes
    delta -= hours * 3600;
    const minutes = Math.floor(delta / 60) % 60;

    if (years < 0 || days < 0 || hours < 0 || minutes < 0) return 'now';
    if (years) return years + 'y ' + days + 'd';
    if (days) return days + 'd ' + hours + 'h';
    if (hours) return hours + 'h ' + minutes + 'm';
    if (minutes) return minutes + 'm';
    return 'now';
  };

  const fieldsDropdown = { instance, setSshModal, fileManager, extended };

  return (
    <>
      <div
        className={`w-full flex items-center ${
          extended ? 'justify-end sm:justify-between' : 'justify-end'
        }`}
      >
        {extended && (
          <div
            className={`flex justify-between items-center ${
              viewMode === WorkspaceRole.manager
                ? 'lg:w-2/5 xl:w-7/12 2xl:w-1/2'
                : 'lg:w-1/3 xl:w-1/2'
            }`}
          >
            <RowInstanceActionsExtended
              setSshModal={setSshModal}
              time={getTime()}
              viewMode={viewMode}
              instance={instance}
            />
            <Text className="hidden lg:block" strong>
              {getTime()}
            </Text>
          </div>
        )}
        <div
          className={`flex justify-end items-center gap-2 w-full ${
            viewMode === WorkspaceRole.manager
              ? 'lg:w-3/5 xl:w-5/12 2xl:w-1/2'
              : 'lg:w-2/3 xl:w-1/2'
          } ${extended ? 'pr-2' : ''}`}
        >
          {!extended && <RowInstanceActionsDropdown {...fieldsDropdown} />}
          {persistent && (
            <RowInstanceActionsPersistent
              instance={instance}
              extended={extended}
            />
          )}
          <RowInstanceActionsDefault
            setSshModal={setSshModal}
            extended={extended}
            instance={instance}
            viewMode={viewMode}
          />
          {extended && <RowInstanceActionsDropdown {...fieldsDropdown} />}
        </div>
      </div>
      <Modal
        title="SSH Connection"
        width={550}
        open={sshModal}
        onOk={() => setSshModal(false)}
        onCancel={() => setSshModal(false)}
        footer={<Button onClick={() => setSshModal(false)}>Close</Button>}
        centered
      >
        <SSHModalContent instanceIp={instance.ip}  hasSSHKeys={hasSSHKeys!} namespace={instance.tenantNamespace} name={instance.name} prettyName={instance.prettyName} onClose={() => setSshModal(false)}/>
      </Modal>
    </>
  );
};

export default RowInstanceActions;
