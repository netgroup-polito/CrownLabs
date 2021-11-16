import { Spin } from 'antd';
import { useContext, useEffect, useState } from 'react';
import { AuthContext } from '../../../contexts/AuthContext';
import {
  useApplyTenantMutation,
  useSshKeysQuery,
} from '../../../generated-types';
import { getTenantPatchJson } from '../../../graphql-components/utils';
import UserPanel from '../UserPanel';
import UserPanelContainer from '../UserPanelContainer/UserPanelContainer';

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
  const [publicKeys, setPublicKeys] = useState<string[]>([]);

  const [applyTenantMutation] = useApplyTenantMutation();

  const { data, loading, error } = useSshKeysQuery({
    variables: { tenantId: userId ?? '' },
    notifyOnNetworkStatusChange: true,
    fetchPolicy: 'network-only',
  });

  useEffect(() => {
    if (!loading) {
      setPublicKeys((data?.tenant?.spec?.publicKeys as string[]) || []);
    }
  }, [loading, data]);

  const tenantSpec = data?.tenant?.spec;

  const updateKeys = async (
    key: { name: string; key: string },
    // TODO: switch to generalized enum
    action: 'ADD' | 'REMOVE'
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
        },
      });
      setPublicKeys(newKeys);
    } catch (error) {
      return false;
    }
    return true;
  };

  return !loading && data && !error ? (
    <UserPanelContainer>
      <UserPanel
        firstName={tenantSpec?.firstName!}
        lastName={tenantSpec?.lastName!}
        email={tenantSpec?.email!}
        username={userId!}
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
