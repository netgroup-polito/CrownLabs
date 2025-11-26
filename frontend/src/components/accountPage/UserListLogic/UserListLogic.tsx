import { type FC, useState, useEffect, useRef, useContext } from 'react';
import { Spin } from 'antd';
import {
  useTenantsQuery,
  useApplyTenantMutation,
  useApplyTenantJsonPatchJsonMutation,
} from '../../../generated-types';
import { getTenantPatchJson, removeWorkspaceJsonPatch } from '../../../graphql-components/utils';
import UserList from '../UserList/UserList';
import {
  makeRandomDigits,
  type UserAccountPage,
  type Workspace,
  type WorkspaceEntry,
} from '../../../utils';
import { Role } from '../../../generated-types';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import {
  ErrorTypes,
  type EnrichedError,
  type SupportedError,
} from '../../../errorHandling/utils';
import { AuthContext } from '../../../contexts/AuthContext';

export interface IUserListLogicProps {
  workspace: Workspace;
}
const UserListLogic: FC<IUserListLogicProps> = props => {
  const { apolloErrorCatcher, makeErrorCatcher } = useContext(ErrorContext);
  const genericErrorCatcher = makeErrorCatcher(ErrorTypes.GenericError);

  const { userId } = useContext(AuthContext);
  const { workspace } = props;
  const [loadingSpinner, setLoadingSpinner] = useState(false);
  const [errors, setErrors] = useState<EnrichedError[]>([]);
  // Used to handle stop while uploading users from CSV
  const [abortUploading, setAbortUploading] = useState<boolean>(false);
  const abortUploadingRef = useRef(false);
  const [uploadedNumber, setUploadedNumber] = useState<number>(0);
  const [uploadedUserNumber, setUploadedUserNumber] = useState(0);
  const [users, setUsers] = useState<UserAccountPage[]>([]);
  const { data, loading, error, refetch } = useTenantsQuery({
    variables: {
      labels: `crownlabs.polito.it/${workspace.namespace}`,
      retrieveWorkspaces: true,
    },
    onError: apolloErrorCatcher,
    notifyOnNetworkStatusChange: true,
  });

  useEffect(() => {
    abortUploadingRef.current = abortUploading;
  }, [abortUploading]);

  const getManager = () => {
    return `${workspace.name}-${userId || makeRandomDigits(10)}`;
  };

  const refreshUserList = async () => await refetch();

  const handleAbort = (value: boolean) => setAbortUploading(value);
  useEffect(() => {
    if (!loading) {
      setUsers(
        data?.tenants?.items?.map(user => ({
          key: user?.metadata?.name || '',
          userid: user?.metadata?.name || '',
          name: user?.spec?.firstName || '',
          surname: user?.spec?.lastName || '',
          email: user?.spec?.email || '',
          currentRole: user?.spec?.workspaces?.find(
            roles => roles?.name === workspace.name,
          )?.role,
          workspaces:
            user?.spec?.workspaces?.map(workspace => ({
              role: workspace?.role as Role,
              name: workspace?.name as string,
            })) || [],
        })) || [],
      );
    }
  }, [loading, data, workspace.name]);

  const [applyTenantMutation] = useApplyTenantMutation();
  const [applyTenantJsonPatchJsonMutation] = useApplyTenantJsonPatchJsonMutation();

  const updateUser = async (user: UserAccountPage, newRole: Role) => {
    try {
      setLoadingSpinner(true);
      
      // Get all workspaces and update only the current one
      const allWorkspaces = users
        .find(u => u.userid === user.userid)!
        .workspaces?.map(({ name, role }) => ({
          name,
          role: name === workspace.name ? newRole : role,
        })) || [];
      
      await applyTenantMutation({
        variables: {
          tenantId: user.userid,
          patchJson: getTenantPatchJson({ workspaces: allWorkspaces }),
          manager: getManager(),
        },
        onError: apolloErrorCatcher,
      });
      setUsers(
        users.map(u => {
          if (u.userid === user.userid) {
            if (u.currentRole === Role.Candidate && workspace.waitingTenants) {
              workspace.waitingTenants--;
              if (workspace.waitingTenants === 0) {
                workspace.waitingTenants = undefined;
              }
            }
            return {
              ...u,
              currentRole: newRole,
              workspaces: allWorkspaces,
            };
          } else {
            return u;
          }
        }),
      );
    } catch (error) {
      genericErrorCatcher(error as SupportedError);
      setLoadingSpinner(false);
      return false;
    }
    setLoadingSpinner(false);
    return true;
  };

  const addUser = async (
    usersToAdd: UserAccountPage[],
    workspaces: WorkspaceEntry[],
  ) => {
    try {
      setLoadingSpinner(true);
      setUploadedNumber(0);
      setUploadedUserNumber(0);
      setErrors([]);
      const usersAdded: UserAccountPage[] = [];

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
                user.userid,
              ),
            },
            onError: apolloErrorCatcher,
          });
          user.workspaces?.push(...workspaces);
          console.log(user.workspaces);
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

  const deleteUser = async (user: UserAccountPage) => {
    try {
      setLoadingSpinner(true);
      // get all the workspaces from the user
      console.log(users.find(u => u.userid === user.userid));
      const workspaces = users.find(u => u.userid === user.userid)!.workspaces;
      // get the index of the workspace to delete
      if(!workspaces) {
        genericErrorCatcher(new Error(`User ${user.userid} has no workspaces`));
        setLoadingSpinner(false);
        return false;
      }
      const workspaceIndex = workspaces.findIndex(w => w.name === workspace.name);
      console.log("Deleting workspace at index:", workspaceIndex);
      console.log(removeWorkspaceJsonPatch(workspaceIndex, workspace.name, workspace.role));
      
      await applyTenantJsonPatchJsonMutation({
        variables: {
          tenantId: user.userid,
          patchJson: removeWorkspaceJsonPatch(workspaceIndex, workspace.name, workspaces[workspaceIndex].role),
          manager: getManager(),
        },
        onError: apolloErrorCatcher,
      });
      
      
      // Refresh user list in local state
      setUsers(users.filter(u => u.userid !== user.userid));
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
        onDeleteUser={deleteUser}
        onUpdateUser={updateUser}
        workspaceNamespace={workspace.namespace}
        workspaceName={workspace.name}
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
