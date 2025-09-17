import type { FC } from 'react';
import { useContext, useMemo, useCallback } from 'react';
import { Spin } from 'antd';
import ActiveView from '../ActiveView/ActiveView';
import { WorkspaceRole } from '../../../utils';
import { TenantContext } from '../../../contexts/TenantContext';
import { makeWorkspace } from '../../../utilsLogic';
import { useOwnedInstancesQuery } from '../../../generated-types';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import type { ApolloError } from '@apollo/client';

const ActiveViewLogic: FC = () => {
  const { apolloErrorCatcher } = useContext(ErrorContext);

  const {
    data: tenantData,
    loading: tenantLoading,
    error: tenantError,
  } = useContext(TenantContext);

  const tenantNs = 'tenant-' + tenantData?.tenant?.metadata?.name;
  const tenantNamespace = tenantData?.tenant?.status?.personalNamespace?.name;

  // Fetch instance data for quota calculations
  const {
    data: instancesData,
    loading: instancesLoading,
    refetch: refetchInstances,
  } = useOwnedInstancesQuery({
    variables: { tenantNamespace: tenantNs || '' },
    skip: !tenantNs,
    onError: apolloErrorCatcher,
    fetchPolicy: 'cache-and-network',
  });

  const workspaces =
    tenantData?.tenant?.spec?.workspaces?.map(makeWorkspace) || [];

  const managerWorkspaces = workspaces?.filter(
    ws => ws.role === WorkspaceRole.manager,
  );

  const tenantId = tenantData?.tenant?.metadata?.name;

  // Calculate quota data
  const quotaData = useMemo(() => {
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
    console.log('Refreshing quota data...'); // Debug log
    try {
      await refetchInstances();
      console.log('Quota data refreshed successfully'); // Debug log
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
      quotaData={{
        consumedQuota: quotaData.consumedQuota,
        workspaceQuota: quotaData.workspaceQuota,
        availableQuota: quotaData.availableQuota,
        showQuotaDisplay: true,
        refreshQuota,
      }} // Pass quota data to ActiveView
    />
  ) : (
    <div className="h-full w-full flex justify-center items-center">
      <Spin size="large" />
    </div>
  );
};

export default ActiveViewLogic;
