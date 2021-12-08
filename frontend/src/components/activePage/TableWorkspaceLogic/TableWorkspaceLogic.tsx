import { FetchPolicy } from '@apollo/client';
import { Spin } from 'antd';
import {
  Dispatch,
  FC,
  SetStateAction,
  useContext,
  useEffect,
  useState,
} from 'react';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { ErrorTypes } from '../../../errorHandling/utils';
import {
  InstancesLabelSelectorQuery,
  UpdatedInstancesLabelSelectorDocument,
  UpdatedInstancesLabelSelectorSubscriptionResult,
  useInstancesLabelSelectorQuery,
} from '../../../generated-types';
import { matchK8sObject, replaceK8sObject } from '../../../k8sUtils';
import { multiStringIncludes, User, WorkspaceRole } from '../../../utils';
import {
  getManagerInstances,
  getSubObjType,
  getTemplatesMapped,
  getWorkspacesMapped,
  notifyStatus,
  SubObjType,
} from '../../../utilsLogic';
import TableWorkspace from '../TableWorkspace/TableWorkspace';

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

  const labels = `crownlabs.polito.it/workspace in (${workspaces
    .map(({ id }) => id)
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
    onCompleted: setDataInstances,
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

          const { instance, updateType } = data?.updateInstanceLabelSelector;
          const { namespace: ns } = instance.metadata!;
          let notify = false;
          const matchNS = ns === tenantNamespace;

          if (prev.instanceList?.instances) {
            let instances = [...prev.instanceList.instances];
            const found = instances.find(matchK8sObject(instance, false));
            const objType = getSubObjType(found!, instance, updateType!);

            switch (objType) {
              case SubObjType.Deletion:
                instances = instances.filter(matchK8sObject(instance, true));
                notify = true;
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
            prev.instanceList.instances = [...instances];
          }

          if (notify && matchNS)
            notifyStatus(instance.status?.phase!, instance, updateType!);

          const newItem = { ...prev };
          setDataInstances(newItem);
          return prev;
        },
      });
      return unsubscribe;
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
