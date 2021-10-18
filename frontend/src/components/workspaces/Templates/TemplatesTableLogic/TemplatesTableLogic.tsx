/* eslint-disable @typescript-eslint/no-unused-vars */
import { FetchPolicy } from '@apollo/client';
import { notification, Spin } from 'antd';
import Button from 'antd-button-color';
import { useContext, useEffect, useState } from 'react';
import { FC } from 'react';
import { AuthContext } from '../../../../contexts/AuthContext';
import {
  useCreateInstanceMutation,
  useDeleteTemplateMutation,
  useOwnedInstancesQuery,
  useWorkspaceTemplatesQuery,
  UpdatedOwnedInstancesSubscriptionResult,
  OwnedInstancesQuery,
  WorkspaceTemplatesQuery,
  UpdatedWorkspaceTemplatesSubscriptionResult,
} from '../../../../generated-types';
import {
  updatedOwnedInstances,
  updatedWorkspaceTemplates,
} from '../../../../graphql-components/subscription';
import { Template, VmStatus, WorkspaceRole } from '../../../../utils';
import { TemplatesEmpty } from '../TemplatesEmpty';
import { TemplatesTable } from '../TemplatesTable';

export interface ITemplateTableLogicProps {
  tenantNamespace: string;
  workspaceNamespace: string;
  role: WorkspaceRole;
}

const fetchPolicy_networkOnly: FetchPolicy = 'network-only';

const TemplatesTableLogic: FC<ITemplateTableLogicProps> = ({ ...props }) => {
  const { userId } = useContext(AuthContext);
  const { tenantNamespace, workspaceNamespace, role } = props;

  const [dataInstances, setDataInstances] = useState<OwnedInstancesQuery>();

  const {
    loading: loadingInstances,
    error: errorInstances,
    subscribeToMore: subscribeToMoreInstances,
  } = useOwnedInstancesQuery({
    variables: { tenantNamespace },
    onCompleted: setDataInstances,
    fetchPolicy: fetchPolicy_networkOnly,
  });

  useEffect(() => {
    if (!loadingInstances) {
      subscribeToMoreInstances({
        document: updatedOwnedInstances,
        variables: {
          tenantNamespace,
        },
        updateQuery: (prev, { subscriptionData }) => {
          const {
            data,
          } = subscriptionData as UpdatedOwnedInstancesSubscriptionResult;

          if (!data?.updateInstance?.instance) return prev;

          const { instance } = data?.updateInstance;

          if (prev.instanceList?.instances) {
            let { instances } = prev.instanceList;
            if (
              instances.find(i => i?.metadata?.name === instance.metadata?.name)
            ) {
              instances = instances.map(i =>
                i?.metadata?.name === instance.metadata?.name ? instance : i
              );
            } else {
              instances = [...instances, instance];
            }
            prev.instanceList.instances = instances;
          }

          const instancePhase = instance.status?.phase;
          if (instancePhase === 'VmiReady') {
            notification.success({
              message:
                instance.spec?.templateCrownlabsPolitoItTemplateRef
                  ?.templateWrapper?.itPolitoCrownlabsV1alpha2Template?.spec
                  ?.templateName,
              description: `Instance started`,
              btn: instance.status?.url && (
                <Button
                  type="success"
                  size="small"
                  onClick={() => window.open(instance.status?.url!, '_blank')}
                >
                  Connect
                </Button>
              ),
            });
          }

          const newItem = { ...prev };
          setDataInstances(newItem);
          return newItem;
        },
      });
    }
  }, [loadingInstances, subscribeToMoreInstances, tenantNamespace, userId]);

  const [dataTemplate, setDataTemplate] = useState<WorkspaceTemplatesQuery>();

  const {
    loading: loadingTemplate,
    error: errorTemplate,
    subscribeToMore: subscribeToMoreTemplates,
  } = useWorkspaceTemplatesQuery({
    variables: { workspaceNamespace },
    onCompleted: setDataTemplate,
    fetchPolicy: fetchPolicy_networkOnly,
  });

  const { instances } = dataInstances?.instanceList ?? {};

  useEffect(() => {
    if (!loadingTemplate) {
      subscribeToMoreTemplates({
        document: updatedWorkspaceTemplates,
        variables: { workspaceNamespace: `${workspaceNamespace}` },
        updateQuery: (prev, { subscriptionData }) => {
          let newData = { ...prev };
          const {
            data,
          } = subscriptionData as UpdatedWorkspaceTemplatesSubscriptionResult;
          const newItem = data?.updatedTemplate?.template!;

          if (!newItem) return prev;

          if (prev.templateList?.templates) {
            const oldItem = prev.templateList.templates.find(
              t => t?.metadata?.id === newItem?.metadata?.id
            );
            if (oldItem) {
              if (JSON.stringify(oldItem) === JSON.stringify(newItem)) {
                //template have been deleted
                newData.templateList!.templates = prev.templateList?.templates!.filter(
                  t => newItem?.metadata?.id !== t?.metadata?.id
                );
              } else {
                //template have been modified
                newData.templateList!.templates = prev.templateList?.templates!.map(
                  t =>
                    newItem?.metadata?.id === t?.metadata?.id ? newItem! : t!
                );
              }
            } else {
              newData.templateList!.templates = [
                ...prev.templateList.templates,
                newItem!,
              ].sort(
                (a, b) => a?.metadata?.id!.localeCompare(b?.metadata?.id!)!
              );
            }
          }

          setDataTemplate(newData);
          return newData;
        },
      });
    }
  }, [loadingTemplate, subscribeToMoreTemplates, userId, workspaceNamespace]);

  const templates: Template[] = (dataTemplate?.templateList?.templates ?? [])
    .map(t => {
      const { spec, metadata } = t!;
      const [environment] = spec?.environmentList!;
      return {
        instances: instances
          ?.filter(
            x =>
              x?.spec?.templateCrownlabsPolitoItTemplateRef?.name ===
              metadata?.id!
          )
          .map((i, n) => {
            return {
              id: n,
              name: `${i?.spec?.prettyName ?? i?.metadata?.name}`,
              ip: i?.status?.ip!,
              status: i?.status?.phase! as VmStatus,
              url: i?.status?.url!,
              gui: i?.spec?.templateCrownlabsPolitoItTemplateRef
                ?.templateWrapper?.itPolitoCrownlabsV1alpha2Template?.spec
                ?.environmentList![0]?.guiEnabled!,
            };
          })!,
        id: t?.metadata?.id!,
        name: t?.spec?.name!,
        gui: !!environment?.guiEnabled!,
        persistent: environment?.persistent!,
        resources: {
          cpu: environment?.resources?.cpu!,
          // TODO: properly handle resources quantities
          memory: environment?.resources?.memory!,
          disk: environment?.resources?.disk!,
        },
      };
    })
    .filter(t => t);

  const [createInstanceMutation] = useCreateInstanceMutation();
  const [
    deleteTemplateMutation,
    { loading: loadingDeleteTemplateMutation },
  ] = useDeleteTemplateMutation();

  return (
    <Spin size="large" spinning={loadingTemplate || loadingInstances}>
      {!loadingTemplate &&
      !loadingInstances &&
      !errorTemplate &&
      !errorInstances &&
      dataInstances &&
      templates &&
      instances ? (
        <TemplatesTable
          tenantNamespace={tenantNamespace}
          workspaceNamespace={workspaceNamespace}
          templates={templates}
          role={role}
          deleteTemplate={(templateId: string) =>
            deleteTemplateMutation({
              variables: {
                workspaceNamespace,
                templateId,
              },
            })
          }
          deleteTemplateLoading={loadingDeleteTemplateMutation}
          editTemplate={() => null}
          createInstance={(templateId: string) =>
            createInstanceMutation({
              variables: {
                templateId,
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
