import { type FC, useContext, useState, useCallback, useMemo } from 'react';
import { Modal, Typography } from 'antd';
import { Button } from 'antd';
import {
  formatRelativeDate,
  type Instance,
  WorkspaceRole,
} from '../../../../utils';
import RowInstanceActionsPersistent from './RowInstanceActionsPersistent';
import RowInstanceActionsDropdown from './RowInstanceActionsDropdown';
import RowInstanceActionsExtended from './RowInstanceActionsExtended';
import SSHModalContent from '../SSHModalContent/SSHModalContent';
import NativeVNCViewer from '../../NativeVNCViewer/NativeVNCViewer';
import RowInstanceActionsDefault from './RowInstanceActionsDefault';
import { PublicExposureModal } from '../PublicExposureModal/PublicExposureModal';
import {
  OwnedInstancesContext,
  type IQuota,
} from '../../../../contexts/OwnedInstancesContext';

const { Text } = Typography;
export interface IRowInstanceActionsProps {
  instance: Instance;
  now: Date;
  fileManager?: boolean;
  hasSSHKeys?: boolean;
  extended: boolean;
  viewMode: WorkspaceRole;
}

const EMPTY_QUOTA: IQuota = {
  instances: 0,
  cpu: 0,
  memory: 0,
  disk: 0,
};

// Returns a human readable string of the time elapsed between now and timeStr
const formatElapsedTime = (
  now: Date,
  timeStr?: string,
  fallback = 'unknown',
) => {
  if (!timeStr) return fallback;
  const time = new Date(timeStr);
  if (isNaN(time.getTime())) return fallback;

  let delta = (now.getTime() - time.getTime()) / 1000;
  const years = Math.floor(delta / (86400 * 365));
  delta -= years * (86400 * 365);
  const days = Math.floor(delta / 86400);
  delta -= days * 86400;
  const hours = Math.floor(delta / 3600) % 24;
  delta -= hours * 3600;
  const minutes = Math.floor(delta / 60) % 60;

  if (years < 0 || days < 0 || hours < 0 || minutes < 0) return 'now';
  if (years) return years + 'y ' + days + 'd';
  if (days) return days + 'd ' + hours + 'h';
  if (hours) return hours + 'h ' + minutes + 'm';
  if (minutes) return minutes + 'm';
  return 'now';
};

const RowInstanceActions: FC<IRowInstanceActionsProps> = ({
  instance,
  now,
  fileManager,
  hasSSHKeys,
  extended,
  viewMode,
}) => {
  const { availableQuota } = useContext(OwnedInstancesContext);

  const workspaceAvailableQuota: IQuota =
    availableQuota?.[instance.workspaceName || ''] || EMPTY_QUOTA;

  // Use the value from the template (mapped via GraphQL)
  const allowPublic = instance.allowPublicExposure;

  const { persistent } = instance;

  const [sshModal, setSshModal] = useState(false);
  const [vncModal, setVncModal] = useState(false);
  const [showExposureModal, setShowExposureModal] = useState(false);

  const onEnablePublicExposure = useCallback(
    () => setShowExposureModal(true),
    [],
  );
  const closeSshModal = useCallback(() => setSshModal(false), []);
  const closeVncModal = useCallback(() => setVncModal(false), []);
  const closeExposureModal = useCallback(() => setShowExposureModal(false), []);

  const timeValue = useMemo(
    () => formatElapsedTime(now, instance.timeStamp, 'unknown'),
    [now, instance.timeStamp],
  );
  const lastAccessValue = useMemo(
    () => formatRelativeDate(instance.lastActivity, now),
    [instance.lastActivity, now],
  );

  const nativeVncEnv = instance.environments?.find(env => !env.guiEnabled);
  const vncWsUrl = (() => {
    if (!nativeVncEnv || !instance.url) return undefined;
    const baseUrl = instance.url.endsWith('/')
      ? instance.url.slice(0, -1)
      : instance.url;
    return `${baseUrl}/${nativeVncEnv.name}/`.replace(/^https/, 'wss');
  })();

  const fieldsDropdown = useMemo(
    () => ({
      instance,
      setSshModal,
      fileManager,
      extended,
      ...(allowPublic ? { onEnablePublicExposure } : {}),
    }),
    [instance, fileManager, extended, allowPublic, onEnablePublicExposure],
  );

  return (
    <>
      <div
        className={`w-full flex items-center ${
          extended ? 'justify-end sm:justify-between' : 'justify-end'
        }`}
      >
        {extended && (
          <div
            className={`flex items-center gap-8 ${
              viewMode === WorkspaceRole.manager
                ? 'lg:w-2/5 xl:w-7/12 2xl:w-1/2'
                : 'lg:w-1/2 xl:w-7/12'
            }`}
          >
            <RowInstanceActionsExtended
              setSshModal={setSshModal}
              time={timeValue}
              viewMode={viewMode}
              instance={instance}
            />
            <div className="hidden lg:flex w-16 justify-center">
              <Text strong>{timeValue}</Text>
            </div>
            <div className="hidden lg:flex w-24 justify-center">
              <Text strong>{lastAccessValue}</Text>
            </div>
          </div>
        )}
        <div
          className={`flex justify-end items-center gap-2 w-full ${
            viewMode === WorkspaceRole.manager
              ? 'lg:w-3/5 xl:w-5/12 2xl:w-1/2'
              : 'lg:w-1/2 xl:w-5/12'
          } ${extended ? 'pr-2' : ''}`}
        >
          {!extended && <RowInstanceActionsDropdown {...fieldsDropdown} />}
          {persistent && (
            <RowInstanceActionsPersistent
              instance={instance}
              extended={extended}
              workspaceAvailableQuota={
                viewMode === WorkspaceRole.user
                  ? workspaceAvailableQuota
                  : undefined
              }
            />
          )}
          <RowInstanceActionsDefault
            setSshModal={setSshModal}
            setVncModal={setVncModal}
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
        onOk={closeSshModal}
        onCancel={closeSshModal}
        footer={<Button onClick={closeSshModal}>Close</Button>}
        centered
      >
        <SSHModalContent
          instanceIp={instance.ip}
          hasSSHKeys={hasSSHKeys!}
          namespace={instance.tenantNamespace}
          name={instance.name}
          prettyName={instance.prettyName}
          onClose={closeSshModal}
          environments={instance.environments}
        />
      </Modal>
      <Modal
        title="Native VNC (provisional test)"
        width={900}
        open={vncModal}
        onOk={closeVncModal}
        onCancel={closeVncModal}
        footer={<Button onClick={closeVncModal}>Close</Button>}
        centered
      >
        {vncWsUrl ? (
          <div style={{ height: '70vh' }}>
            <NativeVNCViewer wsUrl={vncWsUrl} />
          </div>
        ) : (
          <div>
            <p>VNC environment not found or instance URL is missing.</p>
          </div>
        )}
      </Modal>

      {/* show exposure modal only when allowed or in dev */}
      {allowPublic && (
        <PublicExposureModal
          open={showExposureModal}
          onCancel={closeExposureModal}
          allowPublicExposure={allowPublic}
          existingExposure={instance.publicExposure}
          instanceId={instance.name}
          instancePrettyName={instance.prettyName || instance.name}
          tenantNamespace={instance.tenantNamespace}
        />
      )}
    </>
  );
};

export default RowInstanceActions;
