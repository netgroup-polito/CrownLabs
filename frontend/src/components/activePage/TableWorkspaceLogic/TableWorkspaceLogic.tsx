import { FC, useState, useEffect } from 'react';
import { Spin } from 'antd';
import TableWorkspace from '../TableWorkspace/TableWorkspace';
import { WorkspaceRole } from '../../../utils';
import {
  UpdatedInstancesLabelSelectorSubscriptionResult,
  useInstancesLabelSelectorQuery,
  InstancesLabelSelectorQuery,
  UpdatedInstancesLabelSelectorDocument,
  UpdateType,
} from '../../../generated-types';
import { FetchPolicy } from '@apollo/client';
import {
  getManagerInstances,
  notifyStatus,
  filterId,
  getTemplatesMapped,
  getWorkspacesMapped,
} from '../ActiveUtils';
import { matchK8sObject, replaceK8sObject } from '../../../k8sUtils';

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
  const { workspaces, userId, tenantNamespace, filter } = props;
  const [dataInstances, setDataInstances] =
    useState<InstancesLabelSelectorQuery>();

  const label = `crownlabs.polito.it/workspace in (${workspaces
    .map(({ id }) => id)
    .join(',')})`;

  const {
    loading: loadingInstances,
    error: errorInstances,
    subscribeToMore: subscribeToMoreInstances,
  } = useInstancesLabelSelectorQuery({
    variables: {
      labels: label,
    },
    onCompleted: setDataInstances,
    fetchPolicy: fetchPolicy_networkOnly,
  });

  useEffect(() => {
    if (!loadingInstances) {
      subscribeToMoreInstances({
        document: UpdatedInstancesLabelSelectorDocument,
        variables: {
          tenantNamespace,
        },
        updateQuery: (prev, { subscriptionData }) => {
          const { data } =
            subscriptionData as UpdatedInstancesLabelSelectorSubscriptionResult;

          if (!data?.updateInstanceLabelSelector?.instance) return prev;

          const { instance, updateType } = data?.updateInstanceLabelSelector;

          if (prev.instanceList?.instances) {
            let instances = [...prev.instanceList.instances];
            if (updateType === UpdateType.Deleted) {
              instances = instances.filter(matchK8sObject(instance, true));
            } else {
              if (instances.find(matchK8sObject(instance, false))) {
                instances = instances.map(replaceK8sObject(instance));
              } else {
                instances = [...instances, instance];
              }
            }
            prev.instanceList.instances = [...instances];
          }

          notifyStatus(
            instance.status?.phase!,
            instance,
            updateType!,
            tenantNamespace,
            WorkspaceRole.manager
          );

          const newItem = { ...prev };
          setDataInstances(newItem);
          return prev;
        },
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [loadingInstances, subscribeToMoreInstances, tenantNamespace, userId]);

  const instancesMapped = dataInstances?.instanceList?.instances
    ?.map(getManagerInstances)
    .filter(instance => filterId(instance, filter));

  const templatesMapped = getTemplatesMapped(instancesMapped!);

  const workspacesMapped = getWorkspacesMapped(templatesMapped, workspaces);

  return !loadingInstances && !errorInstances && templatesMapped ? (
    <TableWorkspace workspaces={workspacesMapped} />
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
