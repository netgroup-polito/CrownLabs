import { Spin } from 'antd';
import { useContext } from 'react';
import { AuthContext } from '../../../contexts/AuthContext';
import { useTenantQuery } from '../../../generated-types';
import Dashboard from '../Dashboard/Dashboard';

function DashboardLogic() {
  const { userId } = useContext(AuthContext);

  const { data, loading, error } = useTenantQuery({
    variables: { tenantId: userId ?? '' },
    notifyOnNetworkStatusChange: true,
  });

  //startPolling(20000);

  return !loading && data && !error ? (
    <>
      <Dashboard
        tenantNamespace={data.tenant?.status?.personalNamespace?.name!}
        workspaces={
          data.tenant?.spec?.workspaces?.map(workspace => {
            return {
              workspaceId: workspace?.workspaceRef?.workspaceWrapper
                ?.itPolitoCrownlabsV1alpha1Workspace?.spec
                ?.workspaceName as string,
              role: workspace?.role!,
              workspaceNamespace: workspace?.workspaceRef?.workspaceWrapper
                ?.itPolitoCrownlabsV1alpha1Workspace?.status?.namespace
                ?.workspaceNamespace!,
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
}

export default DashboardLogic;
