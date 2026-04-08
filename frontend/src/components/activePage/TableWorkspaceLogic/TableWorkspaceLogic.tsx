import type { FetchPolicy } from '@apollo/client';
import { Spin } from 'antd';
import type { Dispatch, FC, SetStateAction } from 'react';
import { useContext, useEffect, useMemo, useState } from 'react';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { ErrorTypes } from '../../../errorHandling/utils';
import type { UpdatedInstancesLabelSelectorSubscription } from '../../../generated-types';
import {
  UpdatedInstancesLabelSelectorDocument,
  useInstancesLabelSelectorQuery,
} from '../../../generated-types';
import type { User, Workspace } from '../../../utils';
import { multiStringIncludes } from '../../../utils';
import {
  makeGuiInstance,
  getTemplatesMapped,
  getWorkspacesMapped,
} from '../../../utilsLogic';
import TableWorkspace from '../TableWorkspace/TableWorkspace';
import { TenantContext } from '../../../contexts/TenantContext';
import { createInstanceUpdateQuery } from '../../../utils/instanceSubscriptionHandler';

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
    data: instList,
  } = useInstancesLabelSelectorQuery({
    variables: { labels },
    onError: apolloErrorCatcher,
    fetchPolicy: fetchPolicy_networkOnly,
    nextFetchPolicy: 'cache-only',
  });

  useEffect(() => {
    if (!loadingInstances && !errorInstances && !errorsQueue.length) {
      const unsubscribe =
        subscribeToMoreInstances<UpdatedInstancesLabelSelectorSubscription>({
          onError: makeErrorCatcher(ErrorTypes.GenericError),
          document: UpdatedInstancesLabelSelectorDocument,
          variables: { labels },
          updateQuery: createInstanceUpdateQuery({
            tenantNamespace,
            notifier,
          }),
        });
      return unsubscribe;
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [loadingInstances, subscribeToMoreInstances, tenantNamespace, tenantId]);

  const instancesMapped = useMemo(
    () => instList?.instanceList?.instances.map(makeGuiInstance) || [],
    [instList],
  );

  const workspacesMapped = useMemo(() => {
    const instancesFiltered = instancesMapped.filter(
      instance =>
        multiStringIncludes(
          filter,
          instance.prettyName!,
          instance.tenantId!,
          instance.tenantDisplayName!,
        ) || selectiveDestroy.includes(instance.id),
    );

    const templatesMapped = getTemplatesMapped(instancesFiltered, sortingData);

    return getWorkspacesMapped(templatesMapped, workspaces);
  }, [instancesMapped, sortingData, workspaces, filter, selectiveDestroy]);

  return !loadingInstances && !errorInstances ? (
    <TableWorkspace
      instances={instancesMapped}
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
