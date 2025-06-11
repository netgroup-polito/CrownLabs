import type { FetchPolicy } from '@apollo/client';
import { Spin } from 'antd';
import type { Dispatch, FC, SetStateAction } from 'react';
import { useContext, useEffect, useMemo, useState } from 'react';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { ErrorTypes } from '../../../errorHandling/utils';
import type {
  InstancesLabelSelectorQuery,
  UpdatedInstancesLabelSelectorSubscriptionResult,
} from '../../../generated-types';
import {
  UpdatedInstancesLabelSelectorDocument,
  useInstancesLabelSelectorQuery,
} from '../../../generated-types';
import { matchK8sObject, replaceK8sObject } from '../../../k8sUtils';
import type { User, Workspace } from '../../../utils';
import { JSONDeepCopy, multiStringIncludes } from '../../../utils';
import {
  getManagerInstances,
  getSubObjTypeK8s,
  getTemplatesMapped,
  getWorkspacesMapped,
  notifyStatus,
  SubObjType,
} from '../../../utilsLogic';
import TableWorkspace from '../TableWorkspace/TableWorkspace';
import { TenantContext } from '../../../contexts/TenantContext';

const fetchPolicy_networkOnly: FetchPolicy = 'network-only';

export interface ITableWorkspaceLogicProps {
  user: User;
  workspaces: Array<Workspace>;
  filter: string;
  showAdvanced: boolean;
  showCheckbox: boolean;
  collapseAll: boolean;
  expandAll: boolean;
  destroySelectedTrigger: boolean;
  setCollapseAll: Dispatch<SetStateAction<boolean>>;
  setExpandAll: Dispatch<SetStateAction<boolean>>;
  setDestroySelectedTrigger: Dispatch<SetStateAction<boolean>>;
  setSelectedPersistent: Dispatch<SetStateAction<boolean>>;
  selectiveDestroy: string[];
  selectToDestroy: (instanceId: string) => void;
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
    showCheckbox,
    destroySelectedTrigger,
    setDestroySelectedTrigger,
    selectToDestroy,
    selectiveDestroy,
    setSelectedPersistent,
  } = props;

  const { notify: notifier } = useContext(TenantContext);

  const { tenantId, tenantNamespace } = user;
  const [sortingData, setSortingData] = useState<
    Array<{
      sortingType: string;
      sorting: number;
      sortingTemplate: string;
    }>
  >([]);

  const [dataInstances, setDataInstances] = useState<
    NonNullable<InstancesLabelSelectorQuery['instanceList']>['instances']
  >([]);

  const handleManagerSorting = (
    sortingType: string,
    sorting: number,
    sortingTemplate: string,
  ) => {
    const old = sortingData.filter(d => d.sortingTemplate !== sortingTemplate);
    setSortingData([...old, { sortingTemplate, sorting, sortingType }]);
  };

  const labels = `crownlabs.polito.it/workspace in (${workspaces
    .map(({ name }) => name)
    .join(',')})`;

  const { makeErrorCatcher, apolloErrorCatcher, errorsQueue } =
    useContext(ErrorContext);
  const {
    loading: loadingInstances,
    error: errorInstances,
    subscribeToMore: subscribeToMoreInstances,
  } = useInstancesLabelSelectorQuery({
    variables: {
      labels,
    },
    onCompleted: o =>
      setDataInstances(JSONDeepCopy(o.instanceList?.instances || [])),
    onError: apolloErrorCatcher,
    fetchPolicy: fetchPolicy_networkOnly,
  });

  useEffect(() => {
    if (!loadingInstances && !errorInstances && !errorsQueue.length) {
      const unsubscribe = subscribeToMoreInstances({
        onError: makeErrorCatcher(ErrorTypes.GenericError),
        document: UpdatedInstancesLabelSelectorDocument,
        variables: { labels },
        updateQuery: (prev, { subscriptionData }) => {
          const { data } =
            subscriptionData as UpdatedInstancesLabelSelectorSubscriptionResult;
          if (!data?.updateInstanceLabelSelector?.instance) return prev;

          const { instance, updateType } = data.updateInstanceLabelSelector;
          if (!instance.metadata) return prev;

          const { namespace: ns } = instance.metadata;
          let notify = false;
          let objType;
          const matchNS = ns === tenantNamespace;

          setDataInstances(instances => {
            const found = instances.find(matchK8sObject(instance, false));
            objType = getSubObjTypeK8s(found, instance, updateType);

            switch (objType) {
              case SubObjType.Deletion:
                instances = instances.filter(matchK8sObject(instance, true));
                notify = false;
                break;
              case SubObjType.Addition:
                instances = [...instances, instance];
                notify = true;
                break;
              case SubObjType.PrettyName:
                instances = instances.map(replaceK8sObject(instance));
                notify = false;
                break;
              case SubObjType.UpdatedInfo:
                instances = instances.map(replaceK8sObject(instance));
                notify = true;
                break;
              case SubObjType.Drop:
                notify = false;
                break;
              default:
                break;
            }

            if (notify && matchNS) {
              notifyStatus(
                instance.status?.phase,
                instance,
                updateType,
                notifier,
              );
            }
            return instances;
          });

          return prev;
        },
      });
      return unsubscribe;
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [loadingInstances, subscribeToMoreInstances, tenantNamespace, tenantId]);

  const instancesMapped = useMemo(
    () => dataInstances.map(getManagerInstances),
    [dataInstances],
  );

  const instancesFiltered = useMemo(
    () =>
      instancesMapped.filter(
        instance =>
          multiStringIncludes(
            filter,
            instance.prettyName!,
            instance.tenantId!,
            instance.tenantDisplayName!,
          ) || selectiveDestroy.includes(instance.id),
      ),
    [filter, instancesMapped, selectiveDestroy],
  );

  const templatesMapped = useMemo(
    () => getTemplatesMapped(instancesFiltered, sortingData),
    [instancesFiltered, sortingData],
  );

  const workspacesMapped = useMemo(
    () => getWorkspacesMapped(templatesMapped, workspaces),
    [templatesMapped, workspaces],
  );

  return !loadingInstances && !errorInstances ? (
    <TableWorkspace
      instances={instancesMapped!}
      workspaces={workspacesMapped}
      collapseAll={collapseAll}
      expandAll={expandAll}
      setCollapseAll={setCollapseAll}
      setExpandAll={setExpandAll}
      showAdvanced={showAdvanced}
      showCheckbox={showCheckbox}
      handleManagerSorting={handleManagerSorting}
      destroySelectedTrigger={destroySelectedTrigger}
      setDestroySelectedTrigger={setDestroySelectedTrigger}
      selectiveDestroy={selectiveDestroy}
      selectToDestroy={selectToDestroy}
      setSelectedPersistent={setSelectedPersistent}
    />
  ) : (
    <div className="flex justify-center h-full items-center mt-16">
      {loadingInstances ? (
        <Spin size="large" spinning={loadingInstances} />
      ) : (
        <>{errorInstances && <p>{errorInstances.message}</p>}</>
      )}
    </div>
  );
};

export default TableWorkspaceLogic;
