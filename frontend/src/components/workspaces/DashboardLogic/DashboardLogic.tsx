import { Spin } from 'antd';
import type { FC } from 'react';
import { useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { TenantContext } from '../../../contexts/TenantContext';
import { OwnedInstancesContext } from '../../../contexts/OwnedInstancesContext';
import { makeWorkspace } from '../../../utilsLogic';
import Dashboard from '../Dashboard/Dashboard';
import {
  Role,
  TenantsDocument,
  useWorkspacesQuery,
} from '../../../generated-types';
import type { Workspace } from '../../../utils';
import { WorkspaceRole } from '../../../utils';
import { useApolloClient } from '@apollo/client';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { LocalValue, StorageKeys } from '../../../utilsStorage';
import { useQuotaCalculations } from '../QuotaDisplay/useQuotaCalculation';

const dashboard = new LocalValue(StorageKeys.Dashboard_LoadCandidates, 'false');

const DashboardLogic: FC = () => {
  const { apolloErrorCatcher } = useContext(ErrorContext);

  const {
    data: tenantData,
    error: tenantError,
    loading: tenantLoading,
  } = useContext(TenantContext);

  const ws = useMemo(() => {
    return (
      tenantData?.tenant?.spec?.workspaces
        ?.filter(w => w?.role !== Role.Candidate)
        ?.map(makeWorkspace) ?? []
    );
  }, [tenantData?.tenant?.spec?.workspaces]);

  const tenantNs = tenantData?.tenant?.status?.personalNamespace?.name;

  // Get all instances from context
  const { rawInstances, loading: instancesLoading } = useContext(
    OwnedInstancesContext,
  );

  // Use the centralized quota calculation hook
  const quotaCalculations = useQuotaCalculations(
    rawInstances as Parameters<typeof useQuotaCalculations>[0],
    tenantData?.tenant,
  );

  const [viewWs, setViewWs] = useState<Workspace[]>(ws);
  const client = useApolloClient();
  // When templates are created/edited/deleted elsewhere, refetch active queries so UI updates.
  useEffect(() => {
    const handler = (e: Event) => {
      try {
        const detail = (e as CustomEvent)?.detail ?? {};
        console.debug(
          'templatesChanged event received in DashboardLogic',
          detail,
        );
        // Refetch active queries so TemplatesTableLogic / other components update.
        client
          .refetchQueries({ include: 'active' })
          .catch(err =>
            console.warn('refetchQueries failed in DashboardLogic:', err),
          );
      } catch (err) {
        console.warn('templatesChanged handler error in DashboardLogic', err);
      }
    };

    window.addEventListener('templatesChanged', handler as EventListener);
    return () =>
      window.removeEventListener('templatesChanged', handler as EventListener);
  }, [client]);

  const { data: workspaceQueryData } = useWorkspacesQuery({
    variables: {
      labels: 'crownlabs.polito.it/autoenroll=withApproval',
    },
    onError: apolloErrorCatcher,
  });

  const [loadCandidates, setLoadCandidates] = useState(
    dashboard.get() === 'true',
  );

  const wsIsManagedWithApproval = useCallback(
    (w: Workspace): boolean => {
      return (
        w?.role === WorkspaceRole.manager &&
        workspaceQueryData?.workspaces?.items?.find(
          wq => wq?.metadata?.name === w.name,
        ) !== undefined
      );
    },
    [workspaceQueryData?.workspaces?.items],
  );

  useEffect(() => {
    if (loadCandidates) {
      const workspaceQueue: Workspace[] = [];
      const executeNext = () => {
        if (!loadCandidates || workspaceQueue.length === 0) {
          return;
        }
        const w = workspaceQueue.shift();
        client
          .query({
            query: TenantsDocument,
            variables: {
              labels: `crownlabs.polito.it/workspace-${w?.name}=candidate`,
            },
          })
          .then(queryResult => {
            const numCandidate = queryResult.data.tenants.items.length;
            if (numCandidate > 0) {
              ws.find(ws => ws.name === w?.name)!.waitingTenants = numCandidate;
              setViewWs([...ws]);
            }
            executeNext();
          });
      };

      ws
        ?.filter(
          w => w?.role === WorkspaceRole.manager && wsIsManagedWithApproval(w),
        )
        .forEach(w => {
          workspaceQueue.push(w);
          if (workspaceQueue.length === 1) {
            executeNext();
          }
        });
    }
  }, [
    client,
    ws,
    workspaceQueryData?.workspaces?.items,
    loadCandidates,
    wsIsManagedWithApproval,
  ]);

  const selectLoadCandidates = () => {
    if (loadCandidates) {
      ws.forEach(w => (w.waitingTenants = undefined));
    }
    setViewWs([...ws]);
    setLoadCandidates(!loadCandidates);
    dashboard.set(String(!loadCandidates));
  };

  const isLoading = tenantLoading || instancesLoading;

  return !isLoading && tenantData && !tenantError && tenantNs ? (
    <Dashboard
      tenantNamespace={tenantNs}
      tenantPersonalWorkspace={{
        createPWs: tenantData?.tenant?.spec?.quota !== null,
        isPWsCreated:
          tenantData?.tenant?.status?.personalNamespace?.created ?? false,
        quota: quotaCalculations.workspaceQuota
          ? {
              cpu: String(quotaCalculations.workspaceQuota.cpu),
              memory: String(quotaCalculations.workspaceQuota.memory),
              instances: quotaCalculations.workspaceQuota.instances,
            }
          : null,
      }}
      workspaces={viewWs}
      candidatesButton={{
        show: ws.some(w => wsIsManagedWithApproval(w)),
        selected: loadCandidates,
        select: selectLoadCandidates,
      }}
      globalQuota={{
        consumedQuota: quotaCalculations.consumedQuota,
        workspaceQuota: quotaCalculations.workspaceQuota,
        availableQuota: quotaCalculations.availableQuota,
        showQuotaDisplay: true,
      }}
    />
  ) : (
    <div className="h-full w-full flex justify-center items-center">
      <Spin size="large" />
    </div>
  );
};

export default DashboardLogic;
