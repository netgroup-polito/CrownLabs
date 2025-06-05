import { Spin } from 'antd';
import type { FC } from 'react';
import { useEffect, useState, useContext } from 'react';
import { useTenantsQuery } from '../../../generated-types';
import UserListForm from '../UserListForm/UserListForm';
import type { UserAccountPage, WorkspaceEntry } from '../../../utils';
import type { Role } from '../../../generated-types';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import type { SupportedError } from '../../../errorHandling/utils';
import { ErrorTypes } from '../../../errorHandling/utils';

export interface IUserListFormLogicProps {
  onAddUser: (
    newUser: UserAccountPage,
    workspaces: WorkspaceEntry[],
  ) => Promise<boolean>;
  onCancel: () => void;
  workspaceNamespace: string;
  workspaceName: string;
}

const UserListFormLogic: FC<IUserListFormLogicProps> = props => {
  const { onAddUser, onCancel, workspaceNamespace, workspaceName } = props;
  const { apolloErrorCatcher, makeErrorCatcher } = useContext(ErrorContext);
  const genericErrorCatcher = makeErrorCatcher(ErrorTypes.GenericError);

  const [users, setUsers] = useState<UserAccountPage[]>([]);
  const { data, loading, error } = useTenantsQuery({
    variables: {
      labels: `!crownlabs.polito.it/${workspaceNamespace}`,
      retrieveWorkspaces: true,
    },
    onError: apolloErrorCatcher,
    notifyOnNetworkStatusChange: true,
    fetchPolicy: 'no-cache',
  });

  useEffect(() => {
    if (!loading) {
      setUsers(
        data?.tenants?.items
          ?.map(u => u || {})
          .map(({ metadata, spec }) => ({
            key: metadata?.name || '',
            userid: metadata?.name || '',
            name: spec?.firstName || '',
            surname: spec?.lastName || '',
            email: spec?.email || '',
            workspaces:
              spec?.workspaces?.map(w => ({
                role: w?.role as Role,
                name: w?.name || '',
              })) || [],
          })) || [],
      );
    }
  }, [loading, data, workspaceNamespace]);

  const addUser = async (user: UserAccountPage, role: Role) => {
    try {
      user.currentRole = role;
      user.key = user.userid;
      if (await onAddUser(user, [{ name: workspaceName, role: role }])) {
        setUsers(users.filter(u => u.userid !== user.userid));
      }
    } catch (error) {
      genericErrorCatcher(error as SupportedError);
      return false;
    }

    return true;
  };

  return !loading && data && !error ? (
    <>
      <UserListForm users={users} onAddUser={addUser} onCancel={onCancel} />
    </>
  ) : (
    <div className="h-full w-full flex justify-center items-center">
      <Spin size="large" />
    </div>
  );
};

export default UserListFormLogic;
