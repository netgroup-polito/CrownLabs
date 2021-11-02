import { FetchPolicy } from '@apollo/client';
import { FC, useState, useEffect, useContext } from 'react';
import { notification, Spin } from 'antd';
import Button from 'antd-button-color';
import { Instance, WorkspaceRole } from '../../../utils';
import './TableInstance.less';
import TableInstance from './TableInstance';
import { AuthContext } from '../../../contexts/AuthContext';
import {
  useDeleteInstanceMutation,
  useOwnedInstancesQuery,
  UpdatedOwnedInstancesSubscriptionResult,
  OwnedInstancesQuery,
  UpdateType,
} from '../../../generated-types';
import { updatedOwnedInstances } from '../../../graphql-components/subscription';

export interface ITableInstanceLogicProps {
  viewMode: WorkspaceRole;
  showGuiIcon: boolean;
  extended: boolean;
  tenantNamespace: string;
}

const fetchPolicy_networkOnly: FetchPolicy = 'network-only';

const TableInstanceLogic: FC<ITableInstanceLogicProps> = ({ ...props }) => {
  const { tenantNamespace, viewMode, extended, showGuiIcon } = props;
  const { userId } = useContext(AuthContext);
  const [dataInstances, setDataInstances] = useState<OwnedInstancesQuery>();
  const [deleteInstanceMutation] = useDeleteInstanceMutation();

  const startInstance = (idInstance: string, idTemplate: string) => {};
  const stopInstance = (idInstance: string, idTemplate: string) => {};

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
        variables: { tenantNamespace },
        updateQuery: (prev, { subscriptionData }) => {
          const {
            data,
          } = subscriptionData as UpdatedOwnedInstancesSubscriptionResult;

          if (!data?.updateInstance?.instance) return prev;

          const { instance, updateType } = data?.updateInstance;

          if (prev.instanceList?.instances) {
            let instances = prev.instanceList.instances;
            if (updateType === UpdateType.Deleted) {
              instances = instances.filter(i => {
                if (i?.metadata?.name !== instance.metadata?.name) {
                  return true;
                }
                notification.warning({
                  message:
                    i?.spec?.templateCrownlabsPolitoItTemplateRef
                      ?.templateWrapper?.itPolitoCrownlabsV1alpha2Template?.spec
                      ?.templateName,
                  description: `${instance.metadata?.name} deleted`,
                });
                return false;
              });
            } else {
              if (
                instances.find(
                  i => i?.metadata?.name === instance.metadata?.name
                )
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
          if (
            instancePhase === 'VmiReady' &&
            updateType !== UpdateType.Deleted
          ) {
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

  const instances =
    dataInstances?.instanceList?.instances?.map((instance, index) => {
      const { metadata, spec, status } = instance!;
      const {
        environmentList,
        templateName,
      } = spec?.templateCrownlabsPolitoItTemplateRef?.templateWrapper?.itPolitoCrownlabsV1alpha2Template?.spec!;
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
        tenantNamespace: tenantNamespace,
        tenantId: userId,
      } as Instance;
    }) ?? [];
  return !loadingInstances && !errorInstances && dataInstances && instances ? (
    <TableInstance
      showGuiIcon={showGuiIcon}
      viewMode={viewMode}
      instances={instances}
      extended={extended}
      startInstance={startInstance}
      stopInstance={stopInstance}
      destroyInstance={(tenantNamespace: string, instanceId: string) =>
        deleteInstanceMutation({
          variables: { tenantNamespace, instanceId },
        })
      }
    />
  ) : (
    <>
      <div className="flex justify-center h-full items-center">
        {loadingInstances ? (
          <Spin size="large" spinning={loadingInstances} />
        ) : (
          <>{errorInstances && <p>{errorInstances.message}</p>}</>
        )}
      </div>
    </>
  );
};

export default TableInstanceLogic;
