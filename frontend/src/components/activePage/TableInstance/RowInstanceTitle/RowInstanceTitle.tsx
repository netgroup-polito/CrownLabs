import {
  CodeOutlined,
  DesktopOutlined,
  AppstoreAddOutlined,
  DockerOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons';
import { Checkbox, Space, Tooltip, Typography } from 'antd';
import SvgInfinite from '../../../../assets/infinite.svg?react';
import type { ApolloError } from '@apollo/client';
import { type FC, useContext, useEffect, useState } from 'react';
import { ErrorContext } from '../../../../errorHandling/ErrorContext';
import { useApplyInstanceMutation, Phase2 } from '../../../../generated-types';
import { type Instance, WorkspaceRole } from '../../../../utils';
import { setInstancePrettyname } from '../../../../utilsLogic';
import PersistentIcon from '../../../common/PersistentIcon/PersistentIcon';
import RowInstanceStatus from '../RowInstanceStatus/RowInstanceStatus';
import NodeSelectorIcon from '../../../common/NodeSelectorIcon/NodeSelectorIcon';

const { Text } = Typography;
export interface IRowInstanceTitleProps {
  viewMode: WorkspaceRole;
  extended: boolean;
  instance: Instance;
  showGuiIcon: boolean;
  showCheckbox?: boolean;
  selectiveDestroy?: string[];
  selectToDestroy?: (instanceId: string) => void;
}

const RowInstanceTitle: FC<IRowInstanceTitleProps> = ({ ...props }) => {
  const {
    viewMode,
    extended,
    instance,
    showGuiIcon,
    showCheckbox,
    selectiveDestroy,
    selectToDestroy,
  } = props;
  const {
    name,
    prettyName,
    templatePrettyName,
    tenantId,
    tenantDisplayName,
    status,
    persistent,
    nodeSelector,
    gui,
    hasMultipleEnvironments,
  } = instance;

  const [edit, setEdit] = useState(false);
  const [title, setTitle] = useState(prettyName || name);
  const { apolloErrorCatcher } = useContext(ErrorContext);
  const [applyInstanceMutation] = useApplyInstanceMutation({
    onError: apolloErrorCatcher,
  });

  const mutateInstancePrettyname = async (title: string) => {
    setTitle(title);
    try {
      const result = await setInstancePrettyname(
        title,
        instance,
        applyInstanceMutation,
      );
      if (result) setTimeout(setEdit, 400, false);
    } catch (err) {
      apolloErrorCatcher(err as ApolloError);
    }
  };

  const handleEdit = (text: string) => {
    mutateInstancePrettyname(text);
  };

  const cancelEdit = () => {
    setTitle(title);
  };

  useEffect(() => {
    if (prettyName) {
      setTitle(prettyName);
    }
  }, [prettyName]);

  return (
    <>
      <div className="w-full flex justify-start items-center pl-2">
        <Space size="middle">
          {viewMode === WorkspaceRole.manager &&
            selectiveDestroy &&
            selectToDestroy &&
            showCheckbox && (
              <div className="flex mr-2 items-center">
                <Checkbox
                  checked={selectiveDestroy.includes(instance.id)}
                  className="p-0"
                  onClick={() => selectToDestroy(instance.id)}
                />
              </div>
            )}
          <RowInstanceStatus
            status={status}
            environments={instance.environments}
          />

          {viewMode === WorkspaceRole.manager ? (
            <div className="flex items-center gap-4">
              <Text className="w-32">{tenantId}</Text>
              <Text className="hidden w-max lg:w-32 2xl:w-40 md:block" ellipsis>
                {tenantDisplayName}
              </Text>
              <Text
                className="hidden lg:w-32 xl:w-40 2xl:w-max lg:block"
                ellipsis
              >
                {prettyName ?? name}
              </Text>
            </div>
          ) : (
            <>
              {showGuiIcon && extended && (
                <div className="flex items-center">
                  {instance.environments && hasMultipleEnvironments ? (
                    <Tooltip
                      placement="right"
                      title={
                        <div className="p-2">
                          <div className="font-semibold mb-2 text-center">
                            Multiple Environments (
                            {instance.environments.length})
                          </div>
                          {instance.environments.map((env, index) => (
                            <div key={index} className="p-1">
                              <div className="flex items-center gap-2 mb-1">
                                <span className="font-medium">{env.name}</span>
                                {env.guiEnabled ? (
                                  <div className="flex items-center gap-1.5">
                                    <DesktopOutlined
                                      style={{
                                        fontSize: '14px',
                                        color: '#1c7afd',
                                      }}
                                    />
                                    <span className="text-xs">VM GUI</span>
                                    {env.persistent && (
                                      <>
                                        <SvgInfinite
                                          width="14px"
                                          className="success-color-fg ml-1"
                                        />
                                        <span className="text-xs">
                                          Persistent
                                        </span>
                                      </>
                                    )}
                                  </div>
                                ) : env.environmentType === 'Container' ? (
                                  <div className="flex items-center gap-1.5">
                                    <DockerOutlined
                                      style={{
                                        fontSize: '14px',
                                        color: '#1c7afd',
                                      }}
                                    />
                                    <span className="text-xs">
                                      Container SSH
                                    </span>
                                    {env.persistent && (
                                      <>
                                        <SvgInfinite
                                          width="14px"
                                          className="success-color-fg ml-1"
                                        />
                                        <span className="text-xs">
                                          Persistent
                                        </span>
                                      </>
                                    )}
                                  </div>
                                ) : (
                                  <div className="flex items-center gap-1.5">
                                    <CodeOutlined
                                      style={{
                                        fontSize: '14px',
                                        color: '#1c7afd',
                                      }}
                                    />
                                    <span className="text-xs">VM SSH</span>
                                    {env.persistent && (
                                      <>
                                        <SvgInfinite
                                          width="14px"
                                          className="success-color-fg ml-1"
                                        />
                                        <span className="text-xs">
                                          Persistent
                                        </span>
                                      </>
                                    )}
                                  </div>
                                )}
                              </div>
                            </div>
                          ))}
                        </div>
                      }
                    >
                      <AppstoreAddOutlined
                        style={{ fontSize: '24px', color: '#1c7afd' }}
                      />
                    </Tooltip>
                  ) : gui ? (
                    <DesktopOutlined
                      className="primary-color-fg"
                      style={{ fontSize: '24px' }}
                    />
                  ) : (
                    <CodeOutlined
                      className="primary-color-fg"
                      style={{ fontSize: '24px' }}
                    />
                  )}
                </div>
              )}
              <Text
                editable={{
                  tooltip: 'Click to Edit',
                  editing: edit,
                  autoSize: { maxRows: 1 },
                  onChange: value => handleEdit(value),
                  onCancel: cancelEdit,
                }}
                className="w-32 lg:w-40 p-0 m-0"
                onClick={() => setEdit(true)}
                ellipsis
              >
                {title}
              </Text>
              {extended && (
                <Text
                  className="md:w-max hidden xs:block xs:w-28 sm:hidden md:block"
                  ellipsis
                >
                  <i>{templatePrettyName}</i>
                </Text>
              )}
              {persistent && extended && <PersistentIcon />}
              {extended && (() => {
                const stopTimeout = instance.cleanup?.stopAfterInactivity ?? 'never';
                const deleteTimeout = instance.cleanup?.deleteAfterInactivity ?? 'never';
                const hasInactivity = (stopTimeout && stopTimeout !== 'never') || (deleteTimeout && deleteTimeout !== 'never');
                if (!hasInactivity) return null;

                // Parse Go duration string (e.g. "1h", "30m", "24h", "7d") to milliseconds
                const parseDuration = (dur: string): number | null => {
                  if (!dur || dur === 'never') return null;
                  const match = dur.match(/^(\d+)(s|m|h|d)$/);
                  if (!match) return null;
                  const val = parseInt(match[1]);
                  switch (match[2]) {
                    case 's': return val * 1000;
                    case 'm': return val * 60 * 1000;
                    case 'h': return val * 3600 * 1000;
                    case 'd': return val * 86400 * 1000;
                    default: return null;
                  }
                };

                const formatRemaining = (ms: number): string => {
                  if (ms <= 0) return 'imminently';
                  const totalSec = Math.floor(ms / 1000);
                  const days = Math.floor(totalSec / 86400);
                  const hours = Math.floor((totalSec % 86400) / 3600);
                  const minutes = Math.floor((totalSec % 3600) / 60);
                  if (days > 0) return `${days}d ${hours}h`;
                  if (hours > 0) return `${hours}h ${minutes}m`;
                  if (minutes > 0) return `${minutes}m`;
                  return 'less than 1m';
                };

                const now = new Date();
                const isRunning = instance.running;
                const isStopped = instance.status === Phase2.Off;

                // Calculate time remaining until auto-stop (if running)
                let stopRemainingMs: number | null = null;
                if (isRunning && instance.lastActivity && stopTimeout !== 'never') {
                  const stopMs = parseDuration(stopTimeout);
                  if (stopMs) {
                    const lastAct = new Date(instance.lastActivity);
                    stopRemainingMs = (lastAct.getTime() + stopMs) - now.getTime();
                  }
                }

                // Calculate time remaining until deletion (if stopped)
                let deleteRemainingMs: number | null = null;
                if (isStopped && instance.lastPoweredOffTimestamp && deleteTimeout !== 'never') {
                  const deleteMs = parseDuration(deleteTimeout);
                  if (deleteMs) {
                    const lastOff = new Date(instance.lastPoweredOffTimestamp);
                    deleteRemainingMs = (lastOff.getTime() + deleteMs) - now.getTime();
                  }
                }

                // Determine icon urgency color
                const isUrgent = (deleteRemainingMs !== null && deleteRemainingMs < 3600000) ||
                                 (stopRemainingMs !== null && stopRemainingMs < 600000);

                return (
                  <Tooltip
                    title={
                      <div className="text-left">
                        {/* Rules section */}
                        {(stopTimeout !== 'never' || deleteTimeout !== 'never') && (
                          <>
                            This instance will be:<br />
                          </>
                        )}
                        {stopTimeout !== 'never' && (
                          <>
                            ▸ powered off after <b>{stopTimeout}</b> of inactivity<br />
                          </>
                        )}
                        {deleteTimeout !== 'never' && (
                          <>
                            ▸ deleted after being stopped for <b>{deleteTimeout}</b><br />
                          </>
                        )}

                        {/* Running instance status */}
                        {isRunning && stopTimeout !== 'never' && (
                          <>
                            <br />
                            <b>Status:</b> Running<br />
                            {instance.lastActivity ? (
                              <>
                                Last activity: <b>{new Date(instance.lastActivity).toLocaleString()}</b><br />
                                {stopRemainingMs !== null && (
                                  stopRemainingMs > 0
                                    ? <>Auto-stop in: <b style={{ color: stopRemainingMs < 600000 ? '#ff4d4f' : '#faad14' }}>{formatRemaining(stopRemainingMs)}</b></>
                                    : <span style={{ color: '#ff4d4f' }}>⚠ Should have been stopped already</span>
                                )}
                              </>
                            ) : (
                              <>No activity detected yet</>
                            )}
                          </>
                        )}

                        {/* Stopped instance status */}
                        {isStopped && deleteTimeout !== 'never' && (
                          <>
                            <br />
                            <b>Status:</b> Stopped<br />
                            {instance.lastPoweredOffTimestamp ? (
                              <>
                                Stopped since: <b>{new Date(instance.lastPoweredOffTimestamp).toLocaleString()}</b><br />
                                {deleteRemainingMs !== null && (
                                  deleteRemainingMs > 0
                                    ? <>Auto-delete in: <b style={{ color: deleteRemainingMs < 3600000 ? '#ff4d4f' : '#faad14' }}>{formatRemaining(deleteRemainingMs)}</b></>
                                    : <span style={{ color: '#ff4d4f' }}>⚠ Pending deletion</span>
                                )}
                              </>
                            ) : (
                              <>Waiting for powered-off timestamp...</>
                            )}
                          </>
                        )}
                      </div>
                    }
                  >
                    <div className="flex items-center">
                      <ClockCircleOutlined
                        className={isUrgent ? 'ml-1' : 'warning-color-fg ml-1'}
                        style={{ fontSize: '14px', ...(isUrgent ? { color: '#ff4d4f' } : {}) }}
                      />
                    </div>
                  </Tooltip>
                );
              })()}
              {nodeSelector && extended && (
                <NodeSelectorIcon
                  isOnWorkspace={false}
                  nodeSelector={nodeSelector}
                />
              )}
            </>
          )}
        </Space>
      </div>
    </>
  );
};

export default RowInstanceTitle;
