import { FC, useState, useEffect, SetStateAction, Dispatch } from 'react';
import { Spin } from 'antd';
import TableWorkspace from '../TableWorkspace/TableWorkspace';
import { multiStringIncludes, User, WorkspaceRole } from '../../../utils';
import {
  UpdatedInstancesLabelSelectorSubscriptionResult,
  useInstancesLabelSelectorQuery,
  InstancesLabelSelectorQuery,
  UpdatedInstancesLabelSelectorDocument,
  UpdateType,
} from '../../../generated-types';
import { FetchPolicy } from '@apollo/client';
import {
  comparePrettyName,
  matchK8sObject,
  replaceK8sObject,
} from '../../../k8sUtils';
import {
  getManagerInstances,
  getTemplatesMapped,
  getWorkspacesMapped,
  notifyStatus,
} from '../../../utilsLogic';

const fetchPolicy_networkOnly: FetchPolicy = 'network-only';

export interface ITableWorkspaceLogicProps {
  user: User;
  workspaces: Array<{
    prettyName: string;
    role: WorkspaceRole;
    namespace: string;
    id: string;
  }>;
  filter: string;
  collapseAll: boolean;
  expandAll: boolean;
  setCollapseAll: Dispatch<SetStateAction<boolean>>;
  setExpandAll: Dispatch<SetStateAction<boolean>>;
  showAdvanced: boolean;
}

const TableWorkspaceLogic: FC<ITableWorkspaceLogicProps> = ({ ...props }) => {
  const {
    workspaces,
    user,
    filter,
    collapseAll,
    expandAll,
    setExpandAll,
    setCollapseAll,
    showAdvanced,
  } = props;

  const { tenantId, tenantNamespace } = user;
  const [sortingData, setSortingData] = useState<
    Array<{
      sortingType: string;
      sorting: number;
      sortingTemplate: string;
    }>
  >([]);
  const [dataInstances, setDataInstances] =
    useState<InstancesLabelSelectorQuery>();

  const handleManagerSorting = (
    sortingType: string,
    sorting: number,
    sortingTemplate: string
  ) => {
    const old = sortingData.filter(d => d.sortingTemplate !== sortingTemplate);
    setSortingData([...old, { sortingTemplate, sorting, sortingType }]);
  };

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
          let isPrettyNameUpdate = false;

          if (prev.instanceList?.instances) {
            let instances = [...prev.instanceList.instances];
            if (updateType === UpdateType.Deleted) {
              instances = instances.filter(matchK8sObject(instance, true));
            } else {
              const found = instances.find(matchK8sObject(instance, false));
              if (found) {
                isPrettyNameUpdate = !comparePrettyName(found, instance);
                instances = instances.map(replaceK8sObject(instance));
              } else {
                instances = [...instances, instance];
              }
            }
            prev.instanceList.instances = [...instances];
          }

          !isPrettyNameUpdate &&
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
  }, [loadingInstances, subscribeToMoreInstances, tenantNamespace, tenantId]);

  const instancesMapped = dataInstances?.instanceList?.instances
    ?.map(getManagerInstances)
    .filter(instance =>
      multiStringIncludes(
        filter,
        instance.tenantId!,
        instance.tenantDisplayName!
      )
    );

  const templatesMapped = getTemplatesMapped(instancesMapped!, sortingData!);

  const workspacesMapped = getWorkspacesMapped(templatesMapped, workspaces);

  return !loadingInstances && !errorInstances && templatesMapped ? (
    <TableWorkspace
      workspaces={workspacesMapped}
      user={user}
      collapseAll={collapseAll}
      expandAll={expandAll}
      setCollapseAll={setCollapseAll}
      setExpandAll={setExpandAll}
      showAdvanced={showAdvanced}
      handleManagerSorting={handleManagerSorting}
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
