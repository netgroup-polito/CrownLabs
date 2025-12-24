import { useTenantLazyQuery, type TenantQuery } from '../../generated-types';
import Box from '../common/Box';
import TenantSearchForm from './TenantSearchForm';
import { Button, Tooltip } from 'antd';
import { LeftOutlined } from '@ant-design/icons';
import { useState } from 'react';
import TenantPanel from './TenantPanel';

export default function TenantsView() {
  const [tenant, setTenant] = useState<TenantQuery | undefined>(undefined);

  const [loadTenant, { loading }] = useTenantLazyQuery({
    onCompleted: data => setTenant(data),
  });

  const goBack = () => {
    setTenant(undefined);
  };

  return (
    <Box
      header={{
        size: 'middle',
        left: (
          <div className="h-full flex-none flex justify-center items-center w-20">
            {tenant && (
              <Tooltip title="Back">
                <Button
                  type="primary"
                  shape="circle"
                  size="large"
                  icon={<LeftOutlined />}
                  onClick={goBack}
                ></Button>
              </Tooltip>
            )}
          </div>
        ),
        right: (
          <div className="h-full flex-none flex justify-center items-center w-20"></div>
        ),
        center: (
          <div className="h-full flex justify-center items-center px-5">
            <p className="md:text-2xl text-lg text-center mb-0">
              <b>Manage tenant</b>
            </p>
          </div>
        ),
      }}
    >
      <div className="h-full w-full flex justify-center items-center">
        {tenant ? (
          <TenantPanel tenant={tenant} />
        ) : (
          <TenantSearchForm
            isLoading={loading}
            onSearch={tenantId => loadTenant({ variables: { tenantId } })}
          />
        )}
      </div>
    </Box>
  );
}
