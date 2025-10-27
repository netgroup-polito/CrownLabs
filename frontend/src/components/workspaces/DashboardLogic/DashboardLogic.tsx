import { Spin } from 'antd';
import type { FC } from 'react';
import { useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { TenantContext } from '../../../contexts/TenantContext';
import { makeWorkspace } from '../../../utilsLogic';
import Dashboard from '../Dashboard/Dashboard';
import {
  Role,
  TenantsDocument,
  useWorkspacesQuery,
  useOwnedInstancesQuery,
} from '../../../generated-types';
import type { Workspace } from '../../../utils';
import { WorkspaceRole } from '../../../utils';
import { useApolloClient } from '@apollo/client';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { LocalValue, StorageKeys } from '../../../utilsStorage';
import type { ApolloError } from '@apollo/client';

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

  const tenantNs = 'tenant-' + tenantData?.tenant?.metadata?.name;

  // Get all instances for the tenant (includes both workspace and personal instances)
  const {
    data: instancesData,
    loading: instancesLoading,
    refetch: refetchInstances,
  } = useOwnedInstancesQuery({
    variables: { tenantNamespace: tenantNs || '' },
    skip: !tenantNs,
    onError: apolloErrorCatcher,
    fetchPolicy: 'cache-and-network', // Ensure we get fresh data
  });

  // Memoized quota calculations that react to instance changes
  const quotaCalculations = useMemo(() => {
    const parseMemoryToGi = (v: string | number | null | undefined): number => {
      if (v == null) return 0;
      if (typeof v === 'number') return v;
      const s = String(v).trim();
      const m = s.match(/^([\d.]+)\s*(Ki|Mi|Gi|Ti|K|M|G|T)?$/i);
      if (!m) return parseFloat(s.replace(/[^\d.]/g, '')) || 0;
      const val = parseFloat(m[1]);
      const unit = (m[2] || '').toLowerCase();
      const pow = (n: number) => Math.pow(1024, n);
      if (unit === 'ki') return (val * pow(1)) / pow(3);
      if (unit === 'mi') return (val * pow(2)) / pow(3);
      if (unit === 'gi') return val;
      if (unit === 'ti') return (val * pow(4)) / pow(3);
      if (unit === 'k') return (val * 1e3) / pow(3);
      if (unit === 'm') return (val * 1e6) / pow(3);
      if (unit === 'g') return (val * 1e9) / pow(3);
      if (unit === 't') return (val * 1e12) / pow(3);
      return val;
    };

    const consumedQuota = { cpu: 0, memoryGi: 0, instances: 0 };
    const items = instancesData?.instanceList?.instances ?? [];

    for (const inst of items) {
      const resources =
        inst?.spec?.templateCrownlabsPolitoItTemplateRef?.templateWrapper
          ?.itPolitoCrownlabsV1alpha2Template?.spec?.environmentList?.[0]
          ?.resources;
      const cpu = Number(resources?.cpu ?? 0);
      const mem = resources?.memory ?? '0Gi';
      consumedQuota.cpu += cpu;
      consumedQuota.memoryGi += parseMemoryToGi(mem);
      consumedQuota.instances += 1;
    }

    // Calculate available resources from quota - consumed
    const totalQuota = tenantData?.tenant?.status?.quota;
    const availableQuota = {
      cpu:
        (totalQuota?.cpu ? parseFloat(String(totalQuota.cpu)) : 0) -
        consumedQuota.cpu,
      memory: String(
        (totalQuota?.memory ? parseMemoryToGi(totalQuota.memory) : 0) -
          consumedQuota.memoryGi,
      ),
      instances:
        (totalQuota?.instances ? totalQuota.instances : 0) -
        consumedQuota.instances,
    };

    return {
      consumedQuota: {
        cpu: consumedQuota.cpu,
        memory: String(consumedQuota.memoryGi),
        instances: consumedQuota.instances,
      },
      availableQuota,
      workspaceQuota: totalQuota || {
        cpu: 0,
        memory: 0,
        instances: 0,
      },
    };
  }, [
    instancesData?.instanceList?.instances,
    tenantData?.tenant?.status?.quota,
  ]);

  // Enhanced refresh function with better error handling and logging
  const refreshQuota = useCallback(async () => {
    try {
      await refetchInstances();
    } catch (error) {
      console.error('Error refreshing quota data:', error);
      // Type cast the error to ApolloError or create a new one
      if (error && typeof error === 'object' && 'message' in error) {
        apolloErrorCatcher(error as ApolloError);
      } else {
        apolloErrorCatcher(new Error(String(error)) as ApolloError);
      }
    }
  }, [refetchInstances, apolloErrorCatcher]);

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
        createPWs: tenantData?.tenant?.spec?.createPersonalWorkspace ?? false,
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
      // Pass reactive quota data as props
      globalQuota={{
        consumedQuota: quotaCalculations.consumedQuota,
        workspaceQuota: quotaCalculations.workspaceQuota,
        availableQuota: quotaCalculations.availableQuota,
        showQuotaDisplay: true,
        refreshQuota, // Pass the refresh function
      }}
    />
  ) : (
    <div className="h-full w-full flex justify-center items-center">
      <Spin size="large" />
    </div>
  );
};

export default DashboardLogic;
