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
} from '../../../generated-types';
import type { Workspace } from '../../../utils';
import { WorkspaceRole } from '../../../utils';
import { useApolloClient } from '@apollo/client';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { LocalValue, StorageKeys } from '../../../utilsStorage';

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

  const [viewWs, setViewWs] = useState<Workspace[]>(ws);
  const client = useApolloClient();

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

  const tenantNs = tenantData?.tenant?.status?.personalNamespace?.name;

  return !tenantLoading && tenantData && !tenantError && tenantNs ? (
    <>
      <Dashboard
        tenantNamespace={tenantNs}
        workspaces={viewWs}
        candidatesButton={{
          show: ws.some(w => wsIsManagedWithApproval(w)),
          selected: loadCandidates,
          select: selectLoadCandidates,
        }}
      />
    </>
  ) : (
    <div className="h-full w-full flex justify-center items-center">
      <Spin size="large" />
    </div>
  );
};

export default DashboardLogic;
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
  useWorkspaceQuotasQuery,
} from '../../../generated-types';
import type { Workspace } from '../../../utils';
import { WorkspaceRole } from '../../../utils';
import { useApolloClient } from '@apollo/client';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { LocalValue, StorageKeys } from '../../../utilsStorage';

const dashboard = new LocalValue(StorageKeys.Dashboard_LoadCandidates, 'false');

const DashboardLogic: FC = () => {
  const { apolloErrorCatcher } = useContext(ErrorContext);

  const { data: quotasData, loading: quotasLoading, error: quotasError } = useWorkspaceQuotasQuery();

  const workspaceQuotas = useMemo(() => {
    // Map workspace name to quota for easy lookup
    const map: Record<string, { cpu: any; memory: any; instances: number }> = {};
    quotasData?.workspaces?.items?.forEach(ws => {
      if (ws?.metadata?.name && ws?.spec?.quota) {
        map[ws.metadata.name] = ws.spec.quota;
      }
    });
    return map;
  }, [quotasData]); 

  const {
    data: tenantData,
    error: tenantError,
    loading: tenantLoading,
  } = useContext(TenantContext);

  const ws = useMemo(() => {
    const baseWorkspaces =
      tenantData?.tenant?.spec?.workspaces
        ?.filter(w => w?.role !== Role.Candidate)
        ?.map(makeWorkspace) ?? [];

    // Enrich each workspace with its quota, if available
    return baseWorkspaces.map(w => ({
      ...w,
      quota: workspaceQuotas[w.name], // append quota by workspace name
    }));
    // Add workspaceQuotas as a dependency
  }, [tenantData?.tenant?.spec?.workspaces, workspaceQuotas]);

  const [viewWs, setViewWs] = useState<Workspace[]>(ws);
  const client = useApolloClient();

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

  console.log('tenantData:', tenantData);
  console.log('workspaceQuotas:', ws);

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

  const tenantNs = tenantData?.tenant?.status?.personalNamespace?.name;

  return !tenantLoading && tenantData && !tenantError && tenantNs ? (
    <>
      <Dashboard
        tenantNamespace={tenantNs}
        tenantPersonalWorkspace={{
          createPWs: tenantData?.tenant?.spec?.createPersonalWorkspace,
          isPWsCreated: tenantData?.tenant?.status?.personalWorkspace?.created,
          quota: tenantData?.tenant?.status?.quota,
        }}
        workspaces={viewWs}
        candidatesButton={{
          show: ws.some(w => wsIsManagedWithApproval(w)),
          selected: loadCandidates,
          select: selectLoadCandidates,
        }}
      />
    </>
  ) : (
    <div className="h-full w-full flex justify-center items-center">
      <Spin size="large" />
    </div>
  );
};

export default DashboardLogic;
