import type { ApolloError } from '@apollo/client';
import { Alert, Button, List, Popconfirm, Steps, Tag, Typography } from 'antd';
import { type FC, useState } from 'react';
import { useDeleteCrownJobMutation } from '../../../../generated-types';
import {
  apiErrorTexts,
  type CrownJob,
  formatJobTime,
  isJobEnded,
  jobStages,
  PHASE_COLOR,
} from './crownJobUtils';

const { Text } = Typography;

export interface IRunJobListProps {
  tenantNamespace: string;
  jobs: CrownJob[];
  refreshError?: ApolloError;
  onChanged: () => Promise<unknown>;
}

const RunJobList: FC<IRunJobListProps> = ({
  tenantNamespace,
  jobs,
  refreshError,
  onChanged,
}) => {
  const [cancelling, setCancelling] = useState<string | null>(null);
  const [cancelErrors, setCancelErrors] = useState<string[]>([]);
  const [deleteCrownJob] = useDeleteCrownJobMutation();

  const cancel = async (name: string) => {
    setCancelErrors([]);
    setCancelling(name);
    try {
      await deleteCrownJob({ variables: { tenantNamespace, name } });
      await onChanged().catch(() => null);
    } catch (error) {
      setCancelErrors(apiErrorTexts(error as ApolloError));
    } finally {
      setCancelling(null);
    }
  };

  return (
    <>
      {refreshError && (
        <Alert
          className="mb-2"
          type="warning"
          showIcon
          message="The list could not be refreshed:"
          description={apiErrorTexts(refreshError).join(' ')}
        />
      )}
      {cancelErrors.length > 0 && (
        <Alert
          className="mb-2"
          type="error"
          showIcon
          message={cancelErrors.join(' ')}
        />
      )}
      <div className="max-h-96 overflow-y-auto">
        <List
          dataSource={jobs}
          rowKey={job => job.metadata?.name ?? ''}
          locale={{ emptyText: 'No jobs for this instance yet' }}
          renderItem={job => {
            const name = job.metadata?.name ?? '';
            const { phase, reason, message, queuePosition, deadline } =
              job.status ?? {};
            const deleting = Boolean(job.metadata?.deletionTimestamp);
            return (
              <List.Item className="px-0">
                <div className="w-full flex flex-col gap-1 pr-2">
                  <div className="flex items-center justify-between gap-2">
                    <span className="flex flex-wrap items-center gap-2">
                      <Text strong>{name}</Text>
                      <Tag color={phase ? PHASE_COLOR[phase] : undefined}>
                        {phase ?? 'Submitted'}
                      </Tag>
                      {reason && <Text type="secondary">{reason}</Text>}
                      {queuePosition != null && (
                        <Text type="secondary">
                          position in queue: {queuePosition}
                        </Text>
                      )}
                    </span>
                    {!isJobEnded(job) && (
                      <Popconfirm
                        title="Cancel this job?"
                        description="The job is deleted. If it is running, it is stopped."
                        okText="Cancel the job"
                        cancelText="Keep it"
                        disabled={deleting}
                        onConfirm={() => cancel(name)}
                      >
                        <Button
                          danger
                          size="small"
                          shape="round"
                          disabled={deleting}
                          loading={cancelling === name}
                        >
                          {deleting ? 'Deleting' : 'Cancel'}
                        </Button>
                      </Popconfirm>
                    )}
                  </div>
                  {message && <Text>{message}</Text>}
                  <Text type="secondary" className="text-xs">
                    Created {formatJobTime(job.metadata?.creationTimestamp)}
                    {deadline && ` · Deadline ${formatJobTime(deadline)}`}
                  </Text>
                  <Steps
                    className="mt-1"
                    size="small"
                    items={jobStages(phase)}
                  />
                </div>
              </List.Item>
            );
          }}
        />
      </div>
    </>
  );
};

export default RunJobList;
