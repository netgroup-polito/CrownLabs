import { Spin } from 'antd';
import { useContext } from 'react';
import { AuthContext } from '../../../contexts/AuthContext';
import { useTenantQuery } from '../../../generated-types';
import UserPanel from '../UserPanel';
function UserPanelLogic() {
  const { userId } = useContext(AuthContext);

  const { data, loading, error } = useTenantQuery({
    variables: { tenantId: userId ?? '' },
    notifyOnNetworkStatusChange: true,
  });

  //startPolling(20000);

  return !loading && data && !error ? (
    <>
      <UserPanel
        firstName={data.tenant?.spec?.firstName!}
        lastName={data.tenant?.spec?.lastName!}
        email={data.tenant?.spec?.email!}
        username={userId!}
      />
    </>
  ) : (
    <div className="h-full w-full flex justify-center items-center">
      <Spin size="large" />
    </div>
  );
}

export default UserPanelLogic;
