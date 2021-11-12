import { FetchPolicy } from '@apollo/client';
import { FC, useState, useEffect } from 'react';
import { Spin } from 'antd';
import { User, WorkspaceRole } from '../../../utils';
import './TableInstance.less';
import TableInstance from './TableInstance';
import {
  useOwnedInstancesQuery,
  UpdatedOwnedInstancesSubscriptionResult,
  OwnedInstancesQuery,
  UpdateType,
  useSshKeysQuery,
} from '../../../generated-types';
import { updatedOwnedInstances } from '../../../graphql-components/subscription';
import { getInstances, notifyStatus } from '../../../utilsLogic';
import {
  comparePrettyName,
  matchK8sObject,
  replaceK8sObject,
} from '../../../k8sUtils';
export interface ITableInstanceLogicProps {
  viewMode: WorkspaceRole;
  showGuiIcon: boolean;
  extended: boolean;
  user: User;
}

const fetchPolicy_networkOnly: FetchPolicy = 'network-only';

const TableInstanceLogic: FC<ITableInstanceLogicProps> = ({ ...props }) => {
  const { viewMode, extended, showGuiIcon, user } = props;
  const { tenantNamespace, tenantId } = user;
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

  const { data: sshKeysResult } = useSshKeysQuery({
    variables: { tenantId: tenantId ?? '' },
    notifyOnNetworkStatusChange: true,
  });

  const hasSSHKeys = !!sshKeysResult?.tenant?.spec?.publicKeys?.length;

  useEffect(() => {
    if (!loadingInstances) {
      subscribeToMoreInstances({
        document: updatedOwnedInstances,
        variables: { tenantNamespace },
        updateQuery: (prev, { subscriptionData }) => {
          const { data } =
            subscriptionData as UpdatedOwnedInstancesSubscriptionResult;

          if (!data?.updateInstance?.instance) return prev;

          const { instance, updateType } = data?.updateInstance;
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
              WorkspaceRole.user
            );

          const newItem = { ...prev };
          setDataInstances(newItem);
          return newItem;
        },
      });
    }
  }, [loadingInstances, subscribeToMoreInstances, tenantNamespace, tenantId]);

  const instances = dataInstances?.instanceList?.instances?.map((i, n) =>
    getInstances(i!, n, tenantId, tenantNamespace)
  );

  return !loadingInstances && !errorInstances && dataInstances && instances ? (
    <TableInstance
      showGuiIcon={showGuiIcon}
      viewMode={viewMode}
      hasSSHKeys={hasSSHKeys}
      instances={instances}
      extended={extended}
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
