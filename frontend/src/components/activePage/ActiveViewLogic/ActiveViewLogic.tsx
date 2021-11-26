import { FC, useContext } from 'react';
import { Spin } from 'antd';
import ActiveView from '../ActiveView/ActiveView';
import { WorkspaceRole } from '../../../utils';
import { TenantContext } from '../../../graphql-components/tenantContext/TenantContext';

const ActiveViewLogic: FC<{}> = ({ ...props }) => {
  const {
    data: tenantData,
    loading: tenantLoading,
    error: tenantError,
  } = useContext(TenantContext);

  const workspaces =
    tenantData?.tenant?.spec?.workspaces?.map(workspace => {
      const { workspaceWrapperTenantV1alpha2, workspaceId, role } = workspace!;
      const { spec, status } =
        workspaceWrapperTenantV1alpha2?.itPolitoCrownlabsV1alpha1Workspace!;
      return {
        prettyName: spec?.workspaceName as string,
        role: WorkspaceRole[role!],
        namespace: status?.namespace?.workspaceNamespace!,
        id: workspaceId!,
      };
    }) || [];

  const managerWorkspaces = workspaces?.filter(
    ws => ws.role === WorkspaceRole.manager
  );

  return !tenantLoading && tenantData && !tenantError ? (
    <ActiveView
      user={{
        tenantId: tenantData.tenant?.metadata?.tenantId!,
        tenantNamespace: tenantData!.tenant?.status?.personalNamespace?.name!,
      }}
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
