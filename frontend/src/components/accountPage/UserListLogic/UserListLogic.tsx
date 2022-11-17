import { FC, useState, useEffect, useRef, useContext } from 'react';
import { Spin } from 'antd';
import {
  useTenantsQuery,
  useApplyTenantMutation,
} from '../../../generated-types';
import { getTenantPatchJson } from '../../../graphql-components/utils';
import UserList from '../UserList/UserList';
import { makeRandomDigits, UserAccountPage } from '../../../utils';
import { AuthContext } from '../../../contexts/AuthContext';
import { Role } from '../../../generated-types';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { ErrorTypes, SupportedError } from '../../../errorHandling/utils';

export interface IUserListLogicProps {
  workspaceName: string;
  workspaceNamespace: string;
}
const UserListLogic: FC<IUserListLogicProps> = props => {
  const { apolloErrorCatcher, makeErrorCatcher } = useContext(ErrorContext);
  const genericErrorCatcher = makeErrorCatcher(ErrorTypes.GenericError);

  const { userId } = useContext(AuthContext);
  const { workspaceName, workspaceNamespace } = props;
  const [loadingSpinner, setLoadingSpinner] = useState(false);
  const [errors, setErrors] = useState<any[]>([]);
  // Used to handle stop while uploading users from CSV
  const [abortUploading, setAbortUploading] = useState<boolean>(false);
  const abortUploadingRef = useRef(false);
  const [uploadedNumber, setUploadedNumber] = useState<number>(0);
  const [uploadedUserNumber, setUploadedUserNumber] = useState(0);
  const [users, setUsers] = useState<UserAccountPage[]>([]);
  const { data, loading, error, refetch } = useTenantsQuery({
    variables: {
      labels: `crownlabs.polito.it/${workspaceNamespace}`,
      retrieveWorkspaces: true,
    },
    onError: apolloErrorCatcher,
    notifyOnNetworkStatusChange: true,
  });

  useEffect(() => {
    abortUploadingRef.current = abortUploading;
  }, [abortUploading]);

  const getManager = () => {
    return `${workspaceName}-${userId || makeRandomDigits(10)}`;
  };

  const refreshUserList = async () => await refetch();

  const handleAbort = (value: boolean) => setAbortUploading(value);
  useEffect(() => {
    if (!loading) {
      setUsers(
        data?.tenants?.items?.map(user => ({
          key: user?.metadata?.name!,
          userid: user?.metadata?.name!,
          name: user?.spec?.firstName!,
          surname: user?.spec?.lastName!,
          email: user?.spec?.email!,
          currentRole: user?.spec?.workspaces?.find(roles =>
            workspaceName.includes(roles?.name!)
          )?.role!,
          workspaces:
            user?.spec?.workspaces?.map(workspace => ({
              role: workspace?.role! as Role,
              name: workspace?.name! as string,
            })) || [],
        })) || []
      );
    }
  }, [loading, data, workspaceName]);

  const [applyTenantMutation] = useApplyTenantMutation();

  const updateUser = async (user: UserAccountPage, newRole: Role) => {
    try {
      let workspaces = users
        .find(u => u.userid === user.userid)!
        .workspaces?.filter(w => w.name === workspaceName)
        .map(({ name }) => ({ name, role: newRole }));
      setLoadingSpinner(true);
      await applyTenantMutation({
        variables: {
          tenantId: user.userid,
          patchJson: getTenantPatchJson({ workspaces }),
          manager: getManager(),
        },
        onError: apolloErrorCatcher,
      });
      setUsers(
        users.map(u =>
          u.userid === user.userid
            ? {
                ...u,
                currentRole: newRole,
                workspaces,
              }
            : u
        )
      );
    } catch (error) {
      genericErrorCatcher(error as SupportedError);
      setLoadingSpinner(false);
      return false;
    }
    setLoadingSpinner(false);
    return true;
  };

  const addUser = async (usersToAdd: UserAccountPage[], workspaces: any) => {
    try {
      setLoadingSpinner(true);
      setUploadedNumber(0);
      setUploadedUserNumber(0);
      setErrors([]);
      let usersAdded: UserAccountPage[] = [];

      for (const user of usersToAdd) {
        if (abortUploadingRef.current) break;
        try {
          await applyTenantMutation({
            variables: {
              manager: getManager(),
              tenantId: user.userid,
              patchJson: getTenantPatchJson(
                {
                  email: user.email,
                  firstName: user.name,
                  lastName: user.surname,
                  workspaces,
                },
                user.userid
              ),
            },
            onError: apolloErrorCatcher,
          });
          user.workspaces?.push(workspaces);
          setUploadedUserNumber(number => number + 1);
          usersAdded.push(user);
        } catch (error) {
          const enrichedError = {
            ...(error as SupportedError),
            entity: user.userid,
          };
          setErrors(errors => [...errors, enrichedError]);
        }
        setUploadedNumber(number => number + 1);
      }
      setUsers([...users, ...usersAdded]);
    } catch (error) {
      genericErrorCatcher(error as SupportedError);
      setLoadingSpinner(false);
      return false;
    }
    setLoadingSpinner(false);
    return true;
  };

  return !loading && data && !error ? (
    <>
      <UserList
        users={users}
        onAddUser={addUser}
        onUpdateUser={updateUser}
        workspaceNamespace={workspaceNamespace}
        workspaceName={workspaceName}
        uploadedNumber={uploadedNumber}
        uploadedUserNumber={uploadedUserNumber}
        setAbortUploading={handleAbort}
        abortUploading={abortUploading}
        loading={loadingSpinner}
        uploadingErrors={errors}
        genericErrorCatcher={genericErrorCatcher}
        setUploadingErrors={errors => setErrors(errors)}
        refreshUserList={refreshUserList}
      />
    </>
  ) : (
    <div className="h-full w-full flex justify-center items-center">
      <Spin size="large" />
    </div>
  );
};

export default UserListLogic;
