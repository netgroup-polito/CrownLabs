import { useContext, type FC } from 'react';
import type { TenantQuery } from '../../../generated-types';
import { useDeleteTenantMutation } from '../../../generated-types';
import { Button, Popconfirm, message } from 'antd';
import { DeleteOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { ErrorContext } from '../../../errorHandling/ErrorContext';

export interface ITenantInfoProps {
  tenant: TenantQuery;
}

const TenantInfo: FC<ITenantInfoProps> = ({ tenant }) => {
  const navigate = useNavigate();
  const { apolloErrorCatcher } = useContext(ErrorContext);

  const [deleteTenant] = useDeleteTenantMutation({
    onCompleted: () => {
      message.success('Tenant deleted successfully');
      navigate('/tenants');
    },
    onError: apolloErrorCatcher,
  });

  const handleDelete = () => {
    const name = tenant.tenant?.metadata?.name;
    if (name) {
      deleteTenant({ variables: { name } });
    }
  };

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

      <hr className="my-4" />

      <div className="flex justify-end">
        <Popconfirm
          title="Are you sure you want to delete this tenant?"
          onConfirm={handleDelete}
          okText="Yes"
          cancelText="No"
        >
          <Button danger icon={<DeleteOutlined />}>
            Delete Tenant
          </Button>
        </Popconfirm>
      </div>
    </>
  );
};

export default TenantInfo;
