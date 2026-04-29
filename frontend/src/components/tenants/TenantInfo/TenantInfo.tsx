import type { FC } from 'react';
import type { TenantQuery } from '../../../generated-types';

export interface ITenantInfoProps {
  tenant: TenantQuery;
}

const TenantInfo: FC<ITenantInfoProps> = ({ tenant }) => {
  return (
    <>
      <p>
        First name: <strong>{tenant.tenant?.spec?.firstName}</strong>
      </p>
      <p>
        Last name: <strong>{tenant.tenant?.spec?.lastName}</strong>
      </p>
      <p>
        Email: <strong>{tenant.tenant?.spec?.email}</strong>
      </p>

      <hr className="my-4" />

      <p>
        Registration date:
        <strong>
          {tenant.tenant?.metadata?.creationTimestamp
            ? new Date(
                tenant.tenant.metadata.creationTimestamp,
              ).toLocaleDateString()
            : 'N/A'}
        </strong>
      </p>
      <p>
        Last login:{' '}
        <strong>
          {tenant.tenant?.spec?.lastLogin
            ? new Date(tenant.tenant.spec.lastLogin).toLocaleDateString()
            : 'N/A'}
        </strong>
      </p>
    </>
  );
};

export default TenantInfo;
