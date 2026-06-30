import { useMemo } from 'react';
import type { FC } from 'react';
import { Tooltip } from 'antd';
import { ClockCircleOutlined } from '@ant-design/icons';
import { Phase2 } from '../../../generated-types';
import type { Instance } from '../../../utils';

const NEVER = 'never' as const;

const STOP_URGENT_THRESHOLD_MS = 10 * 60 * 1000;   // 10 minutes
const DELETE_URGENT_THRESHOLD_MS = 60 * 60 * 1000;  // 1 hour

const MS_PER_SECOND = 1000;
const MS_PER_MINUTE = 60 * MS_PER_SECOND;
const MS_PER_HOUR = 60 * MS_PER_MINUTE;
const MS_PER_DAY = 24 * MS_PER_HOUR;

type TimeUnit = 's' | 'm' | 'h' | 'd';

const UNIT_TO_MS: Record<TimeUnit, number> = {
  s: MS_PER_SECOND,
  m: MS_PER_MINUTE,
  h: MS_PER_HOUR,
  d: MS_PER_DAY,
};

const parseDuration = (dur: string): number | null => {
  if (!dur || dur === NEVER) return null;
  const match = dur.match(/^(\d+)(s|m|h|d)$/);
  if (!match) return null;
  const val = parseInt(match[1], 10);
  const unit = match[2] as TimeUnit;
  return val * UNIT_TO_MS[unit];
};

const formatRemaining = (ms: number): string => {
  if (ms <= 0) return 'less than 1m';
  const totalSec = Math.floor(ms / MS_PER_SECOND);
  const days = Math.floor(totalSec / (MS_PER_DAY / MS_PER_SECOND));
  const hours = Math.floor((totalSec % (MS_PER_DAY / MS_PER_SECOND)) / 3600);
  const minutes = Math.floor((totalSec % 3600) / 60);
  if (days > 0) return `${days}d ${hours}h`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  if (minutes > 0) return `${minutes}m`;
  return 'less than 1m';
};

type CountdownKind = 'stop' | 'delete' | 'deleteCreation';

const URGENT_ICON_STYLE = { fontSize: '14px', color: '#ff4d4f' };
const NORMAL_ICON_STYLE = { fontSize: '14px' };

export interface IInactivityIconProps {
  instance: Instance;
}

const InactivityIcon: FC<IInactivityIconProps> = ({ instance }) => {
  const stopTimeout = instance.cleanup?.stopAfterInactivity ?? NEVER;
  const deleteTimeout = instance.cleanup?.deleteAfterInactivity ?? NEVER;
  const deleteCreationTimeout = instance.cleanup?.deleteAfterCreation ?? NEVER;

  // It calculates the most imminent event among all active timeouts.
  // The dynamic countdown in the tooltip will only show the single timer
  // for the event that is going to happen first.
  const { targetTime, countdownKind } = useMemo(() => {
    let computedTargetTime: number | null = null;
    let computedCountdownKind: CountdownKind | null = null;

    if (instance.running && stopTimeout !== NEVER && instance.lastActivity) {
      const ms = parseDuration(stopTimeout);
      if (ms) {
        computedTargetTime = new Date(instance.lastActivity).getTime() + ms;
        computedCountdownKind = 'stop';
      }
    } else if (instance.status === Phase2.Off && deleteTimeout !== NEVER && instance.lastPoweredOffTimestamp) {
      const ms = parseDuration(deleteTimeout);
      if (ms) {
        computedTargetTime = new Date(instance.lastPoweredOffTimestamp).getTime() + ms;
        computedCountdownKind = 'delete';
      }
    }

    if (deleteCreationTimeout !== NEVER && instance.timeStamp) {
      const ms = parseDuration(deleteCreationTimeout);
      if (ms) {
        const creationTargetTime = new Date(instance.timeStamp).getTime() + ms;
        if (computedTargetTime === null || creationTargetTime < computedTargetTime) {
          computedTargetTime = creationTargetTime;
          computedCountdownKind = 'deleteCreation';
        }
      }
    }

    return { targetTime: computedTargetTime, countdownKind: computedCountdownKind };
  }, [
    instance.running,
    instance.status,
    instance.lastActivity,
    instance.lastPoweredOffTimestamp,
    instance.timeStamp,
    stopTimeout,
    deleteTimeout,
    deleteCreationTimeout,
  ]);

  const now = Date.now();
  const remainingMs = targetTime !== null ? targetTime - now : null;

  const isUrgent =
    remainingMs !== null &&
    ((countdownKind === 'stop' && remainingMs < STOP_URGENT_THRESHOLD_MS) ||
     (countdownKind === 'delete' && remainingMs < DELETE_URGENT_THRESHOLD_MS) ||
     (countdownKind === 'deleteCreation' && remainingMs < DELETE_URGENT_THRESHOLD_MS));

  const countdownLabel =
    countdownKind === 'stop'
      ? 'Auto-stop for inactivity in'
      : countdownKind === 'delete'
      ? 'Auto-delete for inactivity in'
      : 'Auto-delete for expiration in';

  const tooltipTitle = useMemo(() => (
    <div className="text-left">
      This instance will be:<br />
      {stopTimeout !== NEVER && (
        <>▸ powered off after <b>{stopTimeout}</b> of inactivity<br /></>
      )}
      {deleteTimeout !== NEVER && (
        <>▸ deleted after being stopped for <b>{deleteTimeout}</b><br /></>
      )}
      {deleteCreationTimeout !== NEVER && (
        <>▸ deleted after <b>{deleteCreationTimeout}</b> from creation<br /></>
      )}
      {remainingMs !== null && remainingMs > 0 && (
        <>
          <br />
          {countdownLabel}: <b style={{ color: isUrgent ? '#ff4d4f' : '#faad14' }}>
            {formatRemaining(remainingMs)}
          </b>
        </>
      )}
    </div>
  ), [stopTimeout, deleteTimeout, deleteCreationTimeout, remainingMs, countdownLabel, isUrgent]);

  if (stopTimeout === NEVER && deleteTimeout === NEVER && deleteCreationTimeout === NEVER) return null;

  return (
    <Tooltip title={tooltipTitle}>
      <div className="flex items-center">
        <ClockCircleOutlined
          className={isUrgent ? 'ml-1' : 'warning-color-fg ml-1'}
          style={isUrgent ? URGENT_ICON_STYLE : NORMAL_ICON_STYLE}
        />
      </div>
    </Tooltip>
  );
};

export default InactivityIcon;
