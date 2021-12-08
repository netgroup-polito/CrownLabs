import { Spin } from 'antd';
import { FC, useContext } from 'react';
import { TenantContext } from '../../../graphql-components/tenantContext/TenantContext';
import { makeWorkspace } from '../../../utilsLogic';
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
          tenantData?.tenant?.spec?.workspaces?.map(makeWorkspace) ?? []
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
