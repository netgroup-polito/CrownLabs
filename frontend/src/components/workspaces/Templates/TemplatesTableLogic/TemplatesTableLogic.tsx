/* eslint-disable @typescript-eslint/no-unused-vars */
import { notification, Spin } from 'antd';
import Button from 'antd-button-color';
import {
  Dispatch,
  SetStateAction,
  useContext,
  useEffect,
  useState,
} from 'react';
import { FC } from 'react';
import { AuthContext } from '../../../../contexts/AuthContext';
import {
  useCreateInstanceMutation,
  useOwnedInstancesQuery,
  useWorkspaceTemplatesQuery,
  UpdatedOwnedInstancesSubscriptionResult,
  OwnedInstancesQuery,
} from '../../../../generated-types';
import { updatedOwnedInstances } from '../../../../graphql-components/subscription';
import { Template, VmStatus, WorkspaceRole } from '../../../../utils';
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

  const [dataInstances, setDataInstances] = useState<OwnedInstancesQuery>();

  const {
    loading: loadingInstances,
    error: errorInstances,
    subscribeToMore: subscribeToMoreInstances,
  } = useOwnedInstancesQuery({
    variables: { tenantNamespace },
    onCompleted: setDataInstances,
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
              prev.instanceList.instances = [...instances, instance];
            }
          }

          const instancePhase = instance.status?.phase;
          if (instancePhase === 'VmiReady') {
            notification.success({
              message:
                instance.spec?.templateCrownlabsPolitoItTemplateRef
                  ?.templateWrapper?.itPolitoCrownlabsV1alpha2Template?.spec
                  ?.templateName,
              description: `Instance started`,
              btn: (
                <Button
                  type="success"
                  size="small"
                  onClick={() =>
                    window.open(
                      data?.updateInstance?.instance?.status?.url!,
                      '_blank'
                    )
                  }
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

  const {
    data: dataTemplate,
    loading: loadingTemplate,
    error: errorTemplate,
    refetch: refetchTemplate,
  } = useWorkspaceTemplatesQuery({
    variables: { workspaceNamespace },
    notifyOnNetworkStatusChange: true,
  });

  //This useEffect and the reload state are used to simulate the subscription behaviour, it will be removed in the next PR
  useEffect(() => {
    if (reload && refetchTemplate) {
      setReload(false);
      refetchTemplate({ workspaceNamespace });
    }
  }, [reload, refetchTemplate, setReload, workspaceNamespace]);

  const { instances } = dataInstances?.instanceList ?? {};

  const templates = (dataTemplate?.templateList?.templates ?? [])
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
              name: i?.metadata?.name!,
              ip: i?.status?.ip!,
              status: i?.status?.phase! as VmStatus,
              url: i?.status?.url!,
            };
          })!,
        id: t?.metadata?.id!,
        name: t?.spec?.name!,
        gui: !!environment?.guiEnabled!,
        persistent: environment?.persistent!,
        resources: {
          cpu: environment?.resources?.cpu!,
          // TODO: properly handle resources quantities
          memory: parseInt(environment?.resources?.memory?.split('G')[0]!),
          disk: parseInt(environment?.resources?.disk?.split('G')[0]!),
        },
      };
    })
    .filter(t => t);

  const [createInstanceMutation] = useCreateInstanceMutation();

  return (
    <Spin size="large" spinning={loadingTemplate || loadingInstances}>
      {!loadingTemplate &&
      !loadingInstances &&
      !errorTemplate &&
      !errorInstances &&
      templates &&
      instances ? (
        <TemplatesTable
          tenantNamespace={tenantNamespace}
          workspaceNamespace={workspaceNamespace}
          templates={templates}
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
