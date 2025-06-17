import { Spin } from 'antd';
import { useContext } from 'react';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { useApplyTenantMutation } from '../../../generated-types';
import { TenantContext } from '../../../contexts/TenantContext';
import { getTenantPatchJson } from '../../../graphql-components/utils';
import UserPanel from '../UserPanel';
import UserPanelContainer from '../UserPanelContainer/UserPanelContainer';
import { AuthContext } from '../../../contexts/AuthContext';

const getKeyName = (sshKey: string) => {
  const keyParts = sshKey.split(/\s+/g);
  if (keyParts.length > 2) {
    // Extract from comment part
    keyParts.splice(0, 2); // Remove key-type and key
    if (keyParts.length >= 1) {
      // There is a comment part, rebuild it and extract the name
      const comments = keyParts.join(' ').split(':');
      return comments[comments.length - 1];
    }

    return keyParts.join(' ');
  }
  return null;
};

function UserPanelLogic() {
  const { userId } = useContext(AuthContext);
  const { apolloErrorCatcher } = useContext(ErrorContext);

  const [applyTenantMutation] = useApplyTenantMutation({
    onError: apolloErrorCatcher,
  });

  const { data, error, loading } = useContext(TenantContext);
  const publicKeys = data?.tenant?.spec?.publicKeys?.map(k => k ?? '') ?? [];

  const tenantSpec = data?.tenant?.spec;

  const updateKeys = async (
    key: { name: string; key: string },
    // TODO: switch to generalized enum
    action: 'ADD' | 'REMOVE',
  ) => {
    try {
      const newKeys =
        action === 'ADD'
          ? [...publicKeys, key.key]
          : publicKeys.filter(k => k !== key.key);

      await applyTenantMutation({
        variables: {
          tenantId: userId!,
          patchJson: getTenantPatchJson({
            publicKeys: newKeys,
          }),
          manager: 'frontend-tenant-new-keys',
        },
        onError: apolloErrorCatcher,
      });
    } catch (_error) {
      return false;
    }
    return true;
  };

  return tenantSpec && userId && !loading && data && !error ? (
    <UserPanelContainer>
      <UserPanel
        firstName={tenantSpec.firstName}
        lastName={tenantSpec.lastName}
        email={tenantSpec.email}
        username={userId}
        sshKeys={publicKeys.map((key, i) => ({
          name: getKeyName(key) ?? `Key ${i} `,
          key,
        }))}
        onDeleteKey={key => updateKeys(key, 'REMOVE')}
        onAddKey={key => updateKeys(key, 'ADD')}
      />
    </UserPanelContainer>
  ) : (
    <div className="h-full w-full flex justify-center items-center">
      <Spin size="large" />
    </div>
  );
}

export default UserPanelLogic;
