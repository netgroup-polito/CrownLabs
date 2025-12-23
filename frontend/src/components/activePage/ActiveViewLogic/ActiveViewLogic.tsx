import type { FC } from 'react';
import { useContext, useCallback, useEffect, useRef } from 'react';
import { Spin } from 'antd';
import ActiveView from '../ActiveView/ActiveView';
import { WorkspaceRole } from '../../../utils';
import { TenantContext } from '../../../contexts/TenantContext';
import { OwnedInstancesContext } from '../../../contexts/OwnedInstancesContext';
import { makeWorkspace } from '../../../utilsLogic';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import type { ApolloError } from '@apollo/client';
import { useQuotaCalculations } from '../../workspaces/QuotaDisplay/useQuotaCalculation';
import { useQuotaContext } from '../../../contexts/QuotaContext.types';

const ActiveViewLogic: FC = () => {
  const { apolloErrorCatcher } = useContext(ErrorContext);
  const {
    setConsumedQuota,
    setWorkspaceQuota,
    setAvailableQuota,
    setRefreshQuota,
  } = useQuotaContext();

  const {
    data: tenantData,
    loading: tenantLoading,
    error: tenantError,
  } = useContext(TenantContext);

  const tenantNamespace = tenantData?.tenant?.status?.personalNamespace?.name;

  // Get instance data from context
  const {
    rawInstances,
    loading: instancesLoading,
    refetch: refetchInstances,
  } = useContext(OwnedInstancesContext);

  const workspaces =
    tenantData?.tenant?.spec?.workspaces?.map(makeWorkspace) || [];

  const managerWorkspaces = workspaces?.filter(
    ws => ws.role === WorkspaceRole.manager,
  );

  const tenantId = tenantData?.tenant?.metadata?.name;

  // Use the centralized quota calculation hook with raw instances from context
  const quotaData = useQuotaCalculations(
    rawInstances as Parameters<typeof useQuotaCalculations>[0],
    tenantData?.tenant,
  );

  // Avoid redundant context updates
  const lastAppliedRef = useRef<{
    consumed?: { cpu?: number | string; memory?: string; instances?: number };
    workspace?: { cpu?: number | string; memory?: string; instances?: number };
    available?: { cpu?: number | string; memory?: string; instances?: number };
  }>({});

  useEffect(() => {
    if (!quotaData) return;
    const toConsumed = {
      cpu: quotaData.consumedQuota.cpu,
      memory: String(quotaData.consumedQuota.memory),
      instances: quotaData.consumedQuota.instances,
    };
    const toWorkspace = {
      cpu: quotaData.workspaceQuota.cpu,
      memory: String(quotaData.workspaceQuota.memory),
      instances: quotaData.workspaceQuota.instances,
    };
    const toAvailable = {
      cpu: quotaData.availableQuota.cpu,
      memory: String(quotaData.availableQuota.memory),
      instances: quotaData.availableQuota.instances,
    };

    const eq = (
      a:
        | { cpu?: number | string; memory?: string; instances?: number }
        | undefined,
      b:
        | { cpu?: number | string; memory?: string; instances?: number }
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
  }, [quotaData, setConsumedQuota, setWorkspaceQuota, setAvailableQuota]);

  // register refresh function so UI actions can call refreshQuota?.()
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

  useEffect(() => {
    setRefreshQuota?.(refreshQuota);
    return () => setRefreshQuota?.(undefined);
  }, [refreshQuota, setRefreshQuota]);

  const isLoading = tenantLoading || instancesLoading;

  return !isLoading &&
    tenantData &&
    !tenantError &&
    tenantId &&
    tenantNamespace ? (
    <ActiveView
      user={{ tenantId, tenantNamespace }}
      managerView={managerWorkspaces.length > 0}
      workspaces={managerWorkspaces}
    />
  ) : (
    <div className="h-full w-full flex justify-center items-center">
      <Spin size="large" />
    </div>
  );
};

export default ActiveViewLogic;
