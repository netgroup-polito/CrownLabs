import { Button, Empty, Spin } from 'antd';
import { type FC, useContext, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { ErrorTypes } from '../../../errorHandling/utils';
import {
  type OwnedInstancesQuery,
  type UpdatedOwnedInstancesSubscription,
  useOwnedInstancesQuery,
} from '../../../generated-types';
import { updatedOwnedInstances } from '../../../graphql-components/subscription';
import { TenantContext } from '../../../contexts/TenantContext';
import { matchK8sObject, replaceK8sObject } from '../../../k8sUtils';
import type { WorkspaceRole } from '../../../utils';
import { type Instance, JSONDeepCopy, type User } from '../../../utils';
import {
  getSubObjTypeK8s,
  makeGuiInstance,
  notifyStatus,
  sorter,
  SubObjType,
} from '../../../utilsLogic';
import TableInstance from './TableInstance';
import './TableInstance.less';
export interface ITableInstanceLogicProps {
  viewMode: WorkspaceRole;
  showGuiIcon: boolean;
  extended: boolean;
  user: User;
}

const TableInstanceLogic: FC<ITableInstanceLogicProps> = ({ ...props }) => {
  const { viewMode, extended, showGuiIcon, user } = props;
  const { makeErrorCatcher, apolloErrorCatcher, errorsQueue } =
    useContext(ErrorContext);
  const { tenantNamespace, tenantId } = user;
  const { hasSSHKeys, notify: notifier } = useContext(TenantContext);
  const [dataInstances, setDataInstances] = useState<OwnedInstancesQuery>();
  const [sortingData, setSortingData] = useState<{
    sortingType: string;
    sorting: number;
  }>({ sortingType: '', sorting: 0 });

  const handleSorting = (sortingType: string, sorting: number) => {
    setSortingData({ sortingType, sorting });
  };

  const {
    loading: loadingInstances,
    error: errorInstances,
    subscribeToMore: subscribeToMoreInstances,
  } = useOwnedInstancesQuery({
    skip: !tenantId,
    variables: { tenantNamespace },
    onCompleted: setDataInstances,
    fetchPolicy: 'network-only',
    onError: apolloErrorCatcher,
  });

  useEffect(() => {
    if (!loadingInstances && !errorInstances && !errorsQueue.length) {
      const unsubscribe =
        subscribeToMoreInstances<UpdatedOwnedInstancesSubscription>({
          onError: makeErrorCatcher(ErrorTypes.GenericError),
          document: updatedOwnedInstances,
          variables: { tenantNamespace },
          updateQuery: (prev, { subscriptionData }) => {
            const { data } = subscriptionData;

            if (!data?.updateInstance?.instance) return prev;

            const { instance, updateType } = data.updateInstance;
            let notify = false;
            const newItem = JSONDeepCopy(prev);
            let objType;

            if (newItem.instanceList?.instances) {
              let { instances } = newItem.instanceList;
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
              newItem.instanceList.instances = instances;
            }

            if (notify) {
              notifyStatus(
                instance.status?.phase,
                instance,
                updateType,
                notifier,
              );
            }

            if (objType !== SubObjType.Drop) {
              setDataInstances(newItem);
            }
            return newItem;
          },
        });
      return unsubscribe;
    }
  }, [
    loadingInstances,
    subscribeToMoreInstances,
    tenantNamespace,
    tenantId,
    errorsQueue.length,
    errorInstances,
    apolloErrorCatcher,
    makeErrorCatcher,
    notifier,
  ]);

  const instances =
    dataInstances?.instanceList?.instances
      ?.map(i => makeGuiInstance(i, tenantId))
      .sort((a, b) =>
        sorter(
          a,
          b,
          sortingData.sortingType as keyof Instance,
          sortingData.sorting,
        ),
      ) || [];

  return (
    <>
      {!loadingInstances && !errorInstances && dataInstances ? (
        instances.length ? (
          <TableInstance
            showGuiIcon={showGuiIcon}
            viewMode={viewMode}
            hasSSHKeys={hasSSHKeys}
            instances={instances}
            extended={extended}
            handleSorting={handleSorting}
            showAdvanced={true}
          />
        ) : (
          <div className="w-full h-full flex-grow flex flex-wrap content-center justify-center py-5 ">
            <div className="w-full pb-10 flex justify-center">
              <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={false} />
            </div>
            <p className="text-xl xs:text-3xl text-center px-5 xs:px-24">
              No running instances
            </p>
            <div className="w-full pb-10 flex justify-center">
              <Link to="/">
                <Button type="primary" shape="round" size="large">
                  Create Instance
                </Button>
              </Link>
            </div>
          </div>
        )
      ) : (
        <div className="flex justify-center h-full items-center">
          {loadingInstances ? (
            <Spin size="large" spinning={loadingInstances} />
          ) : (
            <>{errorInstances && <p>{errorInstances.message}</p>}</>
          )}
        </div>
      )}
    </>
  );
};

export default TableInstanceLogic;
