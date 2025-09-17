import type { FC } from 'react';
import { useContext, useMemo } from 'react';
import { Spin } from 'antd';
import ActiveView from '../ActiveView/ActiveView';
import { WorkspaceRole } from '../../../utils';
import { TenantContext } from '../../../contexts/TenantContext';
import { makeWorkspace } from '../../../utilsLogic';

const ActiveViewLogic: FC = () => {
  const {
    data: tenantData,
    loading: tenantLoading,
    error: tenantError,
  } = useContext(TenantContext);

  const workspaces =
    tenantData?.tenant?.spec?.workspaces?.map(makeWorkspace) || [];

  const managerWorkspaces = workspaces?.filter(
    ws => ws.role === WorkspaceRole.manager,
  );

  const tenantId = tenantData?.tenant?.metadata?.name;
  const tenantNamespace = tenantData?.tenant?.status?.personalNamespace?.name;

  // Calculate quota data
  const quotaData = useMemo(() => {
    const totalQuota = tenantData?.tenant?.status?.quota || {
      cpu: 0,
      memory: '0Gi',
      instances: 0,
    };

    const consumedQuota = {
      cpu: 0,
      memory: '0Gi',
      instances: 0,
    };

    return {
      consumedQuota,
      availableQuota: {
        cpu: totalQuota.cpu - consumedQuota.cpu,
        memory: String(
          parseFloat(totalQuota.memory) - parseFloat(consumedQuota.memory),
        ),
        instances: totalQuota.instances - consumedQuota.instances,
      },
      workspaceQuota: totalQuota,
    };
  }, [tenantData?.tenant?.status?.quota]);

  return !tenantLoading &&
    tenantData &&
    !tenantError &&
    tenantId &&
    tenantNamespace ? (
    <ActiveView
      user={{ tenantId, tenantNamespace }}
      managerView={managerWorkspaces.length > 0}
      workspaces={managerWorkspaces}
      quotaData={quotaData} // Pass quota data to ActiveView
    />
  ) : (
    <div className="h-full w-full flex justify-center items-center">
      <Spin size="large" />
    </div>
  );
};

export default ActiveViewLogic;
