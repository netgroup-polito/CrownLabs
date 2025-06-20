import { type FC, useContext, useMemo } from 'react';
import {
  Role,
  useApplyTenantMutation,
  useWorkspacesQuery,
} from '../../../../generated-types';
import { ErrorContext } from '../../../../errorHandling/ErrorContext';
import { Spin } from 'antd';
import { WorkspacesList } from '../WorkspacesList';
import { availableWorkspaces, makeWorkspace } from '../../../../utilsLogic';
import { TenantContext } from '../../../../contexts/TenantContext';
import {
  type WorkspacesAvailable,
  WorkspacesAvailableAction,
} from '../../../../utils';
import { getTenantPatchJson } from '../../../../graphql-components/utils';
import {
  ErrorTypes,
  type SupportedError,
} from '../../../../errorHandling/utils';
import { AuthContext } from '../../../../contexts/AuthContext';

const WorkspaceListLogic: FC = () => {
  const { apolloErrorCatcher, makeErrorCatcher } = useContext(ErrorContext);
  const genericErrorCatcher = makeErrorCatcher(ErrorTypes.GenericError);

  const { userId } = useContext(AuthContext);
  const { data: tenantData } = useContext(TenantContext);

  const { data, loading, error } = useWorkspacesQuery({
    variables: {
      labels: 'crownlabs.polito.it/autoenroll in (immediate, withApproval)',
    },
    onError: apolloErrorCatcher,
  });

  const [applyTenantMutation] = useApplyTenantMutation();

  const availableWs = useMemo(
    () =>
      availableWorkspaces(
        data?.workspaces?.items ?? [],
        tenantData?.tenant?.spec?.workspaces?.map(makeWorkspace) ?? [],
      ),
    [data?.workspaces?.items, tenantData?.tenant?.spec?.workspaces],
  );

  const applyWorkspaces = async (w: { name: string; role: Role }[]) => {
    try {
      await applyTenantMutation({
        variables: {
          tenantId: userId ?? '',
          patchJson: getTenantPatchJson({
            workspaces: w,
          }),
          manager: userId ?? '',
        },
        onError: apolloErrorCatcher,
      });
    } catch (error) {
      genericErrorCatcher(error as SupportedError);
    }
  };

  const getWorkspaces = () => {
    return (tenantData?.tenant?.spec?.workspaces ?? []).map(ws => {
      return {
        name: ws?.name ?? '',
        role: ws?.role ?? Role.User,
      };
    });
  };

  const addWorkspace = (w: WorkspacesAvailable, desiredRole: Role) => {
    const workspaces = getWorkspaces();
    workspaces.push({ name: w.name, role: desiredRole });
    applyWorkspaces(workspaces);
  };

  const action = (w: WorkspacesAvailable) => {
    switch (w.action) {
      case WorkspacesAvailableAction.Join:
        addWorkspace(w, Role.User);
        break;
      case WorkspacesAvailableAction.AskToJoin:
        addWorkspace(w, Role.Candidate);
        break;
      default:
        throw new Error('Action not supported');
    }
  };

  return !loading && data && !error ? (
    <div className="h-full w-full flex justify-center items-center">
      <WorkspacesList
        workspacesAvailable={availableWs}
        action={action}
      ></WorkspacesList>
    </div>
  ) : (
    <div className="h-full w-full flex justify-center items-center">
      <Spin size="large" />
    </div>
  );
};

export default WorkspaceListLogic;
