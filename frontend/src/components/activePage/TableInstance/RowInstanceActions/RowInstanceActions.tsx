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
import RowInstanceActionsDefault from './RowInstanceActionsDefault';
import { PublicExposureModal } from '../PublicExposureModal/PublicExposureModal';
import RunJobModal from '../RunJob/RunJobModal';
import { isRunJobAvailable } from '../RunJob/crownJobUtils';
import { useInstanceCrownJobs } from '../RunJob/useInstanceCrownJobs';
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
  const [showExposureModal, setShowExposureModal] = useState(false);
  const [runJobOpen, setRunJobOpen] = useState(false);

  // Only the user's own view: the manager view shows other users' instances.
  const canRunJob =
    viewMode === WorkspaceRole.user &&
    !fileManager &&
    isRunJobAvailable(instance);
  const crownJobs = useInstanceCrownJobs(instance, {
    skip: !canRunJob,
    open: runJobOpen,
  });
  const { activeCount } = crownJobs;
  const openRunJob = useCallback(() => setRunJobOpen(true), []);
  const closeRunJob = useCallback(() => setRunJobOpen(false), []);

  const onEnablePublicExposure = useCallback(
    () => setShowExposureModal(true),
    [],
  );
  const closeSshModal = useCallback(() => setSshModal(false), []);
  const closeExposureModal = useCallback(() => setShowExposureModal(false), []);

  const timeValue = useMemo(
    () => formatElapsedTime(now, instance.timeStamp, 'unknown'),
    [now, instance.timeStamp],
  );
  const lastAccessValue = useMemo(
    () => formatRelativeDate(instance.lastActivity, now),
    [instance.lastActivity, now],
  );

  const fieldsDropdown = useMemo(
    () => ({
      instance,
      setSshModal,
      fileManager,
      extended,
      ...(allowPublic ? { onEnablePublicExposure } : {}),
      ...(canRunJob ? { onRunJob: openRunJob, activeJobs: activeCount } : {}),
    }),
    [
      instance,
      fileManager,
      extended,
      allowPublic,
      onEnablePublicExposure,
      canRunJob,
      openRunJob,
      activeCount,
    ],
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
          <RowInstanceActionsDropdown {...fieldsDropdown} />
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
            extended={extended}
            instance={instance}
            viewMode={viewMode}
            {...(canRunJob
              ? { onRunJob: openRunJob, activeJobs: activeCount }
              : {})}
          />
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
      {canRunJob && instance.job && (
        <RunJobModal
          instance={instance}
          job={instance.job}
          open={runJobOpen}
          onClose={closeRunJob}
          jobs={crownJobs.jobs}
          usedNames={crownJobs.usedNames}
          refreshError={crownJobs.error}
          refresh={crownJobs.refetch}
        />
      )}
    </>
  );
};

export default RowInstanceActions;
