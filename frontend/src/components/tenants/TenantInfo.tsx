import type { FC } from 'react';
import type { TenantQuery } from '../../generated-types';

export interface ITenantInfoProps {
  tenant: TenantQuery;
}

const TenantInfo: FC<ITenantInfoProps> = ({ tenant }) => {
  return (
    <>
      <h3>Personal</h3>
      <p>
        First name: <strong>{tenant.tenant?.spec?.firstName}</strong>
      </p>
      <p>
        Last name: <strong>{tenant.tenant?.spec?.lastName}</strong>
      </p>
      <p>
        Email: <strong>{tenant.tenant?.spec?.email}</strong>
      </p>
    </>
  );
};

export default TenantInfo;
