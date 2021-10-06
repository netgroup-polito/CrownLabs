import { Spin } from 'antd';
import { useContext } from 'react';
import { AuthContext } from '../../../contexts/AuthContext';
import { useTenantQuery } from '../../../generated-types';
import UserPanel from '../UserPanel';
import UserPanelContainer from '../UserPanelContainer/UserPanelContainer';
function UserPanelLogic() {
  const { userId } = useContext(AuthContext);

  const { data, loading, error } = useTenantQuery({
    variables: { tenantId: userId ?? '' },
    notifyOnNetworkStatusChange: true,
  });

  const tenantSpec = data?.tenant?.spec;

  return !loading && data && !error ? (
    <UserPanelContainer>
      <UserPanel
        firstName={tenantSpec?.firstName!}
        lastName={tenantSpec?.lastName!}
        email={tenantSpec?.email!}
        username={userId!}
      />
    </UserPanelContainer>
  ) : (
    <div className="h-full w-full flex justify-center items-center">
      <Spin size="large" />
    </div>
  );
}

export default UserPanelLogic;
