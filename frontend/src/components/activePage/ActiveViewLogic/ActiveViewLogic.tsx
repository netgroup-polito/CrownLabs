import { FC, useContext, useState, useEffect } from 'react';
import { Spin } from 'antd';
import ActiveView from '../ActiveView/ActiveView';
import { AuthContext } from '../../../contexts/AuthContext';
import { updatedTenant } from '../../../graphql-components/subscription';
import { TenantQuery, useTenantQuery } from '../../../generated-types';
import { User, WorkspaceRole } from '../../../utils';

const ActiveViewLogic: FC<{}> = ({ ...props }) => {
  const { userId } = useContext(AuthContext);
  const [data, setData] = useState<TenantQuery>();
  const [user, setUser] = useState<User>();

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

  useEffect(() => {
    if (!loading && data) {
      setUser({
        tenantId: userId!,
        tenantNamespace: data!.tenant?.status?.personalNamespace?.name!,
      });
    }
  }, [data, loading, userId]);

  const workspaces =
    data?.tenant?.spec?.workspaces?.map(workspace => {
      const { spec, status } =
        workspace?.workspaceRef?.workspaceWrapper
          ?.itPolitoCrownlabsV1alpha1Workspace!;
      return {
        prettyName: spec?.workspaceName as string,
        role: WorkspaceRole[workspace?.role!],
        namespace: status?.namespace?.workspaceNamespace!,
        id: workspace?.workspaceRef?.workspaceId!,
      };
    }) || [];

  const managerWorkspaces = workspaces?.filter(
    ws => ws.role === WorkspaceRole.manager
  );

  return !loading && data && !error && user ? (
    <ActiveView
      user={user}
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
