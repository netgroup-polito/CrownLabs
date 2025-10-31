import type { FC } from 'react';
import { useContext, useCallback } from 'react';
import { Spin } from 'antd';
import ActiveView from '../ActiveView/ActiveView';
import { WorkspaceRole } from '../../../utils';
import { TenantContext } from '../../../contexts/TenantContext';
import { makeWorkspace } from '../../../utilsLogic';
import { useOwnedInstancesQuery } from '../../../generated-types';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import type { ApolloError } from '@apollo/client';
import { useQuotaCalculations } from '../../workspaces/QuotaDisplay/useQuotaCalculation';

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

  // Use the centralized quota calculation hook
  const quotaData = useQuotaCalculations(
    instancesData?.instanceList?.instances?.filter(
      (i): i is NonNullable<typeof i> => i != null,
    ),
    tenantData?.tenant,
  );

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
      }}
    />
  ) : (
    <div className="h-full w-full flex justify-center items-center">
      <Spin size="large" />
    </div>
  );
};

export default ActiveViewLogic;
