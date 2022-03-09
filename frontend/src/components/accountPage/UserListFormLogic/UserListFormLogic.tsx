import { Spin } from 'antd';
import { FC, useEffect, useState, useContext } from 'react';
import { useTenantsQuery } from '../../../generated-types';
import UserListForm from '../UserListForm/UserListForm';
import { UserAccountPage } from '../../../utils';
import { Role } from '../../../generated-types';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { ErrorTypes, SupportedError } from '../../../errorHandling/utils';

export interface IUserListFormLogicProps {
  onAddUser: (newUser: UserAccountPage, workspaces: any[]) => Promise<Boolean>;
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
        data?.tenants?.items?.map(user => ({
          key: user?.metadata?.name!,
          userid: user?.metadata?.name!,
          name: user?.spec?.firstName!,
          surname: user?.spec?.lastName!,
          email: user?.spec?.email!,
          workspaces:
            user?.spec?.workspaces?.map(workspace => ({
              role: workspace?.role as Role,
              name: workspace?.name! as string,
            })) || [],
        })) || []
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
