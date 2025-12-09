import { Spin } from 'antd';
import type { FC } from 'react';
import {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  useRef,
} from 'react';
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
import type { ApolloError } from '@apollo/client';
import { useQuotaCalculations } from '../QuotaDisplay/useQuotaCalculation';
import { useQuotaContext } from '../../../contexts/QuotaContext.types';

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
  const {
    rawInstances,
    loading: instancesLoading,
    refetch: refetchInstances,
  } = useContext(OwnedInstancesContext);

  // Use the centralized quota calculation hook
  const quotaCalculations = useQuotaCalculations(
    rawInstances as Parameters<typeof useQuotaCalculations>[0],
    tenantData?.tenant,
  );

  // push computed quotas into the global QuotaContext so the AppLayout StatusBar (mounted higher) updates
  const {
    setConsumedQuota,
    setWorkspaceQuota,
    setAvailableQuota,
    setRefreshQuota,
  } = useQuotaContext();

  // keep last applied quotas to avoid redundant context updates
  const lastAppliedRef = useRef<{
    consumed?: { cpu: number; memory: string; instances: number };
    workspace?: { cpu: number; memory: string; instances: number };
    available?: { cpu: number; memory: string; instances: number };
  }>({});

  useEffect(() => {
    if (!quotaCalculations) return;
    const toConsumed = {
      cpu: quotaCalculations.consumedQuota.cpu,
      memory: String(quotaCalculations.consumedQuota.memory),
      instances: quotaCalculations.consumedQuota.instances,
    };
    const toWorkspace = {
      cpu: quotaCalculations.workspaceQuota.cpu,
      memory: String(quotaCalculations.workspaceQuota.memory),
      instances: quotaCalculations.workspaceQuota.instances,
    };
    const toAvailable = {
      cpu: quotaCalculations.availableQuota.cpu,
      memory: String(quotaCalculations.availableQuota.memory),
      instances: quotaCalculations.availableQuota.instances,
    };

    const eq = (
      a:
        | { cpu: number | string; memory: string; instances: number }
        | undefined,
      b:
        | { cpu: number | string; memory: string; instances: number }
        | undefined,
    ) =>
      !!a &&
      !!b &&
      a.cpu === b.cpu &&
      a.memory === b.memory &&
      a.instances === b.instances;

    if (!eq(lastAppliedRef.current.consumed, toConsumed)) {
      setConsumedQuota?.(toConsumed);
      lastAppliedRef.current.consumed = toConsumed;
    }
    if (!eq(lastAppliedRef.current.workspace, toWorkspace)) {
      setWorkspaceQuota?.(toWorkspace);
      lastAppliedRef.current.workspace = toWorkspace;
    }
    if (!eq(lastAppliedRef.current.available, toAvailable)) {
      setAvailableQuota?.(toAvailable);
      lastAppliedRef.current.available = toAvailable;
    }
  }, [
    quotaCalculations,
    setConsumedQuota,
    setWorkspaceQuota,
    setAvailableQuota,
  ]);

  // Enhanced refresh function with better error handling and logging
  const refreshQuota = useCallback(async () => {
    try {
      await refetchInstances();
    } catch (error) {
      console.error('Error refreshing quota data:', error);
      if (error && typeof error === 'object' && 'message' in error) {
        apolloErrorCatcher(error as ApolloError);
      } else {
        apolloErrorCatcher(new Error(String(error)) as ApolloError);
      }
    }
  }, [refetchInstances, apolloErrorCatcher]);

  // register the refresh function with the QuotaProvider so other components can call it
  useEffect(() => {
    setRefreshQuota?.(refreshQuota);
    return () => setRefreshQuota?.(undefined);
  }, [refreshQuota, setRefreshQuota]);

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
      globalQuota={{
        consumedQuota: quotaCalculations.consumedQuota,
        workspaceQuota: quotaCalculations.workspaceQuota,
        availableQuota: quotaCalculations.availableQuota,
        showQuotaDisplay: true,
        refreshQuota,
      }}
    />
  ) : (
    <div className="h-full w-full flex justify-center items-center">
      <Spin size="large" />
    </div>
  );
};

export default DashboardLogic;
