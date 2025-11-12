import { type FC, useState, useEffect, useRef, useContext } from 'react';
import { Spin } from 'antd';
import {
  useTenantsQuery,
  useApplyTenantMutation,
  useTenantQuery,
  useReplaceTenantMutation,
} from '../../../generated-types';
import { getTenantPatchJson } from '../../../graphql-components/utils';
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
  const [replaceTenantMutation] = useReplaceTenantMutation();
  const { refetch: refetchTenant } = useTenantQuery({
    variables: { tenantId: '' },
    skip: true,
  });

  const updateUser = async (user: UserAccountPage, newRole: Role) => {
    try {
      const workspaces = users
        .find(u => u.userid === user.userid)!
        .workspaces?.filter(w => w.name === workspace.name)
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
              workspaces,
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

      // Fetch the full tenant data to preserve all fields
      const { data: tenantData } = await refetchTenant({
        tenantId: user.userid,
      });

      if (!tenantData?.tenant) {
        throw new Error('Tenant not found');
      }

      const tenant = tenantData.tenant;

      // Build updated workspaces array (remove current workspace)
      const updatedWorkspaces =
        tenant.spec?.workspaces?.filter(w => w?.name !== workspace.name).map(w => ({
          name: w?.name || '',
          role: w?.role || Role.User,
        })) || [];

      // Build tenant input copying all fields and replacing workspaces
      const tenantInput = {
        apiVersion: tenant.apiVersion,
        kind: tenant.kind,
        metadata: {
          name: tenant.metadata?.name,
          namespace: tenant.metadata?.namespace,
          labels: tenant.metadata?.labels,
          annotations: tenant.metadata?.annotations,
          uid: tenant.metadata?.uid,
          resourceVersion: tenant.metadata?.resourceVersion,
          generation: tenant.metadata?.generation,
          creationTimestamp: tenant.metadata?.creationTimestamp,
          deletionTimestamp: tenant.metadata?.deletionTimestamp,
          deletionGracePeriodSeconds: tenant.metadata?.deletionGracePeriodSeconds,
          finalizers: tenant.metadata?.finalizers,
          selfLink: tenant.metadata?.selfLink,
        },
        spec: {
          email: tenant.spec?.email || '',
          firstName: tenant.spec?.firstName || '',
          lastName: tenant.spec?.lastName || '',
          lastLogin: tenant.spec?.lastLogin ?? undefined,
          createPersonalWorkspace: tenant.spec?.createPersonalWorkspace ?? undefined,
          createSandbox: tenant.spec?.createSandbox ?? undefined,
          publicKeys: tenant.spec?.publicKeys ?? undefined,
          workspaces: updatedWorkspaces,
          quota: tenant.spec?.quota
            ? {
                cpu: tenant.spec.quota.cpu,
                memory: tenant.spec.quota.memory,
                instances: tenant.spec.quota.instances,
              }
            : undefined,
        },
        status: tenant.status
          ? {
              personalNamespace: {
                name: tenant.status.personalNamespace?.name ?? undefined,
                created: tenant.status.personalNamespace?.created ?? false,
              },
              sandboxNamespace: {
                name: tenant.status.sandboxNamespace?.name ?? undefined,
                created: tenant.status.sandboxNamespace?.created ?? false,
              },
              quota: tenant.status.quota
                ? {
                    cpu: tenant.status.quota.cpu,
                    memory: tenant.status.quota.memory,
                    instances: tenant.status.quota.instances,
                  }
                : undefined,
              subscriptions: tenant.status.subscriptions ?? undefined,
              ready: tenant.status.ready ?? false,
            }
          : undefined,
      };
      // Replace the tenant (preserving all fields except the removed workspace)
      
      await replaceTenantMutation({
        variables: {
          tenantId: user.userid,
          manager: getManager(),
          tenantInput,
        },
        onError: apolloErrorCatcher,
      });
      
      // Refresh user list in local state
      await refreshUserList();
    } catch (error) {
      genericErrorCatcher(error as SupportedError);
      setLoadingSpinner(false);
      return false;
    }
    setLoadingSpinner(false);
    return true;
  };
  ;
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
