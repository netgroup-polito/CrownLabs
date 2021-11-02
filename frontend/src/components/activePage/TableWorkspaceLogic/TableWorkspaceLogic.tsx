import { FC, useState, useEffect } from 'react';
import { Spin, notification } from 'antd';
import Button from 'antd-button-color';
import TableWorkspace from '../TableWorkspace/TableWorkspace';
import {
  /* Workspace, */ WorkspaceRole,
  Instance,
  Template,
} from '../../../utils';
import {
  /* WorkspaceTemplatesQuery,
  useWorkspaceTemplatesQuery,
  UpdatedWorkspaceTemplatesSubscriptionResult, */
  UpdatedInstancesLabelSelectorSubscriptionResult,
  useInstancesLabelSelectorQuery,
  InstancesLabelSelectorQuery,
  UpdatedInstancesLabelSelectorDocument,
  UpdateType,
} from '../../../generated-types';
//import { updatedWorkspaceTemplates } from '../../../graphql-components/subscription';
import { FetchPolicy } from '@apollo/client';
//import { updatedInstancesLabelSelector } from '../../../graphql-components/subscription';

const fetchPolicy_networkOnly: FetchPolicy = 'network-only';

export interface ITableWorkspaceLogicProps {
  userId: string;
  tenantNamespace: string;
  workspaces: Array<{
    prettyName: string;
    role: WorkspaceRole;
    namespace: string;
    id: string;
  }>;
  filter: string;
}

const TableWorkspaceLogic: FC<ITableWorkspaceLogicProps> = ({ ...props }) => {
  const { workspaces, userId, tenantNamespace } = props;
  const [templatesMapped, setTemplatesMapped] = useState<Array<Template>>();
  //const [dataTemplate, setDataTemplate] = useState<WorkspaceTemplatesQuery>();
  const [
    dataInstances,
    setDataInstances,
  ] = useState<InstancesLabelSelectorQuery>();

  const {
    loading: loadingInstances,
    error: errorInstances,
    subscribeToMore: subscribeToMoreInstances,
  } = useInstancesLabelSelectorQuery({
    variables: {
      labels: `crownlabs.polito.it/workspace in (${workspaces.map(
        ({ id }) => id
      )})`,
    },
    onCompleted: setDataInstances,
    fetchPolicy: fetchPolicy_networkOnly,
  });

  useEffect(() => {
    subscribeToMoreInstances({
      document: UpdatedInstancesLabelSelectorDocument,
      variables: {
        tenantNamespace,
      },
      updateQuery: (prev, { subscriptionData }) => {
        const {
          data,
        } = subscriptionData as UpdatedInstancesLabelSelectorSubscriptionResult;

        if (!data?.updateInstanceLabelSelector?.instance) return prev;

        const { instance, updateType } = data?.updateInstanceLabelSelector;

        if (prev.instanceList?.instances) {
          let instances = prev.instanceList.instances;
          if (updateType === UpdateType.Deleted) {
            instances = instances.filter(i => {
              if (i?.metadata?.name !== instance.metadata?.name) {
                return true;
              } else {
                notification.warning({
                  message:
                    i?.spec?.templateCrownlabsPolitoItTemplateRef
                      ?.templateWrapper?.itPolitoCrownlabsV1alpha2Template?.spec
                      ?.templateName,
                  description: `${instance.metadata?.name} deleted`,
                });
                return false;
              }
            });
          } else {
            if (
              instances.find(i => i?.metadata?.name === instance.metadata?.name)
            ) {
              instances = instances.map(i =>
                i?.metadata?.name === instance.metadata?.name ? instance : i
              );
            } else {
              instances = [...instances, instance];
            }
          }
          prev.instanceList.instances = instances;
        }

        const instancePhase = instance.status?.phase;
        if (instancePhase === 'VmiReady' && updateType !== UpdateType.Deleted) {
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
  }, [loadingInstances, subscribeToMoreInstances, tenantNamespace, userId]);

  useEffect(() => {
    if (!loadingInstances) {
      const templateList = Array.from(
        new Set(
          dataInstances?.instanceList?.instances?.map(
            i =>
              i?.spec?.templateCrownlabsPolitoItTemplateRef?.templateWrapper
                ?.itPolitoCrownlabsV1alpha2Template?.spec?.templateName
          )
        )
      );
      const instancesMapped = dataInstances?.instanceList?.instances?.map(
        (instance, index) => {
          const { metadata, spec, status } = instance!;
          const {
            environmentList,
            templateName,
          } = spec?.templateCrownlabsPolitoItTemplateRef?.templateWrapper?.itPolitoCrownlabsV1alpha2Template?.spec!;
          const {
            firstName,
            lastName,
          } = spec?.tenantCrownlabsPolitoItTenantRef?.tenantWrapper?.itPolitoCrownlabsV1alpha1Tenant?.spec!;
          return {
            id: index,
            name: metadata?.name,
            gui: environmentList?.[0]?.guiEnabled,
            persistent: environmentList?.[0]?.persistent,
            idTemplate: spec?.templateCrownlabsPolitoItTemplateRef?.name!,
            templatePrettyName: templateName,
            ip: instance?.status?.ip,
            status: instance?.status?.phase,
            url: status?.url,
            timeStamp: metadata?.creationTimestamp,
            tenantId: spec?.tenantCrownlabsPolitoItTenantRef?.tenantId,
            tenantNamespace: `tenant-${spec?.tenantCrownlabsPolitoItTenantRef?.tenantId}`,
            tenantDisplayName: `${firstName}\n${lastName}`,
            workspaceId: spec?.templateCrownlabsPolitoItTemplateRef?.namespace,
          } as Instance;
        }
      );
      const templates = templateList.map(t => {
        const instancesTmp = instancesMapped?.filter(
          ({ templatePrettyName: tpn }) => tpn === t
        );
        return {
          id: instancesTmp![0].idTemplate,
          name: t,
          gui: instancesTmp![0]?.gui,
          persistent: instancesTmp![0]?.persistent,
          resources: { cpu: 0, memory: '', disk: '' },
          instances: instancesTmp,
          workspaceId: instancesTmp![0].workspaceId,
        } as Template;
      });
      setTemplatesMapped(templates);
    }
  }, [
    loadingInstances,
    dataInstances?.instanceList,
    userId,
    errorInstances,
    workspaces,
    dataInstances,
    tenantNamespace,
  ]);

  return !loadingInstances && !errorInstances && templatesMapped ? (
    <TableWorkspace
      workspaces={workspaces.map(ws => ({
        id: ws.id,
        title: ws.prettyName,
        role: ws.role,
        templates: templatesMapped.filter(
          ({ workspaceId: id }) => id!.replace(/workspace-/g, '') === ws.id
        ),
      }))}
      filter={''}
    />
  ) : (
    <div className="flex justify-center h-full items-center">
      {loadingInstances ? (
        <Spin size="large" spinning={loadingInstances} />
      ) : (
        <>{errorInstances && <p>{errorInstances.message}</p>}</>
      )}
    </div>
  );
};

export default TableWorkspaceLogic;
