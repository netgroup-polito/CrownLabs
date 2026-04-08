import type { FC } from 'react';
import { useContext } from 'react';
import { Spin } from 'antd';
import ActiveView from '../ActiveView/ActiveView';
import { WorkspaceRole } from '../../../utils';
import { TenantContext } from '../../../contexts/TenantContext';
import { OwnedInstancesContext } from '../../../contexts/OwnedInstancesContext';
import { makeWorkspace } from '../../../utilsLogic';

const ActiveViewLogic: FC = () => {
  const {
    data: tenantData,
    loading: tenantLoading,
    error: tenantError,
  } = useContext(TenantContext);
  const { loading: instancesLoading } = useContext(OwnedInstancesContext);

  const tenantNamespace = tenantData?.tenant?.status?.personalNamespace?.name;

  const workspaces =
    tenantData?.tenant?.spec?.workspaces?.map(makeWorkspace) || [];

  const managerWorkspaces = workspaces?.filter(
    ws => ws.role === WorkspaceRole.manager,
  );

  const tenantId = tenantData?.tenant?.metadata?.name;

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
