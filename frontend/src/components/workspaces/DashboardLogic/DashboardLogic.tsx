import { Spin } from 'antd';
import { FC, useContext } from 'react';
import { TenantContext } from '../../../graphql-components/tenantContext/TenantContext';
import { WorkspaceRole } from '../../../utils';
import Dashboard from '../Dashboard/Dashboard';

const DashboardLogic: FC<{}> = () => {
  const {
    data: tenantData,
    error: tenantError,
    loading: tenantLoading,
  } = useContext(TenantContext);

  return !tenantLoading && tenantData && !tenantError ? (
    <>
      <Dashboard
        tenantNamespace={tenantData.tenant?.status?.personalNamespace?.name!}
        workspaces={
          tenantData?.tenant?.spec?.workspaces?.map(workspace => {
            return {
              workspaceId: workspace?.workspaceWrapperTenantV1alpha2
                ?.itPolitoCrownlabsV1alpha1Workspace?.spec
                ?.prettyName as string,
              role: WorkspaceRole[workspace?.role!],
              workspaceNamespace:
                workspace?.workspaceWrapperTenantV1alpha2
                  ?.itPolitoCrownlabsV1alpha1Workspace?.status?.namespace
                  ?.name!,
              workspaceName: workspace?.name!,
            };
          }) ?? []
        }
      />
    </>
  ) : (
    <div className="h-full w-full flex justify-center items-center">
      <Spin size="large" />
    </div>
  );
};

export default DashboardLogic;
