import { FetchPolicy } from '@apollo/client';
import { FC, useState, useEffect } from 'react';
import { Spin, Empty } from 'antd';
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
import { Link } from 'react-router-dom';
import Button from 'antd-button-color';
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
    fetchPolicy: 'network-only',
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

  const instances =
    dataInstances?.instanceList?.instances?.map((i, n) =>
      getInstances(i!, n, tenantId, tenantNamespace)
    ) || [];

  return !loadingInstances && !errorInstances && dataInstances ? (
    instances.length ? (
      <TableInstance
        showGuiIcon={showGuiIcon}
        viewMode={viewMode}
        hasSSHKeys={hasSSHKeys}
        instances={instances}
        extended={extended}
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
