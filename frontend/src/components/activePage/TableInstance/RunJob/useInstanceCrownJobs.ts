import { useMemo } from 'react';
import { useCrownJobsQuery } from '../../../../generated-types';
import type { Instance } from '../../../../utils';
import { instanceJobs, isJobEnded, type CrownJob } from './crownJobUtils';

const POLL_OPEN_MS = 2000;
// For the count on the button while the popup is closed.
const POLL_CLOSED_MS = 30000;

const NO_JOBS: (CrownJob | null)[] = [];

export const useInstanceCrownJobs = (
  instance: Instance,
  { skip, open }: { skip: boolean; open: boolean },
) => {
  const { data, previousData, error, refetch } = useCrownJobsQuery({
    variables: { tenantNamespace: instance.tenantNamespace },
    skip,
    pollInterval: skip ? 0 : open ? POLL_OPEN_MS : POLL_CLOSED_MS,
    fetchPolicy: 'cache-and-network',
  });

  // A failed poll has no data: keep showing the last list next to its error.
  const namespaceJobs = (data ?? previousData)?.crownJobList?.jobs ?? NO_JOBS;

  return useMemo(() => {
    const jobs = instanceJobs(namespaceJobs, instance.name);
    return {
      jobs,
      activeCount: jobs.filter(job => !isJobEnded(job)).length,
      usedNames: new Set(namespaceJobs.map(job => job?.metadata?.name ?? '')),
      error,
      refetch,
    };
  }, [namespaceJobs, instance.name, error, refetch]);
};
