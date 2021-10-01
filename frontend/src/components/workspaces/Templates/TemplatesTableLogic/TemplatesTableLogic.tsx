/* eslint-disable @typescript-eslint/no-unused-vars */
import { Spin } from 'antd';
import { Dispatch, SetStateAction, useContext, useEffect } from 'react';
import { FC } from 'react';
import { AuthContext } from '../../../../contexts/AuthContext';
import {
  useCreateInstanceMutation,
  useOwnedInstancesQuery,
  useWorkspaceTemplatesQuery,
} from '../../../../generated-types';
import { VmStatus, WorkspaceRole } from '../../../../utils';
import { TemplatesEmpty } from '../TemplatesEmpty';
import { TemplatesTable } from '../TemplatesTable';

export interface ITemplateTableLogicProps {
  tenantNamespace: string;
  workspaceNamespace: string;
  role: WorkspaceRole;
  reload: boolean;
  setReload: Dispatch<SetStateAction<boolean>>;
}

const TemplatesTableLogic: FC<ITemplateTableLogicProps> = ({ ...props }) => {
  const { userId } = useContext(AuthContext);
  const {
    tenantNamespace,
    workspaceNamespace,
    role,
    reload,
    setReload,
  } = props;

  const {
    data: dataTemplate,
    loading: loadingTemplate,
    error: errorTemplate,
    refetch: refetchTemplate,
  } = useWorkspaceTemplatesQuery({
    variables: { workspaceNamespace },
    notifyOnNetworkStatusChange: true,
  });

  const {
    data: dataInstances,
    loading: loadingInstances,
    error: errorInstances,
    startPolling: startPollingInstances,
  } = useOwnedInstancesQuery({
    variables: { tenantNamespace },
  });

  //This polling is used to simulate the subscription behaviour, it will be removed in the next PR
  startPollingInstances(500);

  //This useEffect and the reload state are used to simulate the subscription behaviour, it will be removed in the next PR
  useEffect(() => {
    if (reload === true && refetchTemplate) {
      setReload(false);
      refetchTemplate({ workspaceNamespace });
    }
  }, [reload, refetchTemplate, setReload, workspaceNamespace]);

  const templates = dataTemplate?.templateList?.templates;

  const [createInstanceMutation] = useCreateInstanceMutation();

  return (
    <Spin size="large" spinning={loadingTemplate || loadingInstances}>
      {!loadingTemplate &&
      !loadingInstances &&
      !errorTemplate &&
      !errorInstances &&
      templates &&
      templates.length ? (
        <TemplatesTable
          tenantNamespace={tenantNamespace}
          workspaceNamespace={workspaceNamespace}
          templates={templates.map(t => {
            const environment =
              t?.spec?.environmentList && t?.spec?.environmentList[0];
            return {
              instances: dataInstances?.instanceList?.instances
                ?.filter(
                  x =>
                    x?.spec?.templateCrownlabsPolitoItTemplateRef?.name ===
                    t?.metadata?.id!
                )
                .map((i, n) => {
                  return {
                    id: n,
                    name: `Instance ${n}`,
                    ip: i?.status?.ip!,
                    status: i?.status?.phase! as VmStatus,
                    url: i?.status?.url!,
                  };
                })!,
              id: t?.metadata?.id!,
              name: t?.spec?.name!,
              gui: (environment && environment.guiEnabled!) || false,
              persistent: environment?.persistent!,
              resources: {
                cpu: environment?.resources?.cpu!,
                memory: parseInt(
                  environment?.resources?.memory?.split('G')[0]!
                ),
                disk: parseInt(environment?.resources?.disk?.split('G')[0]!),
              },
            };
          })}
          role={role}
          deleteTemplate={() => null}
          editTemplate={() => null}
          createInstance={(id: string) =>
            createInstanceMutation({
              variables: {
                templateName: id,
                tenantNamespace,
                tenantId: userId!,
                workspaceNamespace,
              },
            })
          }
        />
      ) : (
        <div
          className={
            loadingTemplate ||
            loadingInstances ||
            errorTemplate ||
            errorInstances
              ? 'invisible'
              : 'visible'
          }
        >
          <TemplatesEmpty role={role} />
        </div>
      )}
    </Spin>
  );
};

export default TemplatesTableLogic;
