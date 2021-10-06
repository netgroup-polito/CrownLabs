import { Spin } from 'antd';
import { FC, useContext, useEffect, useState } from 'react';
import { AuthContext } from '../../../contexts/AuthContext';
import { TenantQuery, useTenantQuery } from '../../../generated-types';

import { updatedTenant } from '../../../graphql-components/subscription';
import Dashboard from '../Dashboard/Dashboard';

const DashboardLogic: FC<{}> = () => {
  const { userId } = useContext(AuthContext);
  const [data, setData] = useState<TenantQuery>();

  const { loading, error, subscribeToMore } = useTenantQuery({
    variables: { tenantId: userId ?? '' },
    onCompleted: setData,
    fetchPolicy: 'network-only',
  });

  useEffect(() => {
    if (!loading) {
      subscribeToMore({
        variables: { tenantId: userId ?? '' },
        document: updatedTenant,
        updateQuery: (prev, { subscriptionData: { data } }) => {
          if (!data) return prev;
          setData(data);
          return data;
        },
      });
    }
  }, [subscribeToMore, loading, userId]);

  return !loading && data && !error ? (
    <>
      <Dashboard
        tenantNamespace={data.tenant?.status?.personalNamespace?.name!}
        workspaces={
          data?.tenant?.spec?.workspaces?.map(workspace => {
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
};

export default DashboardLogic;
