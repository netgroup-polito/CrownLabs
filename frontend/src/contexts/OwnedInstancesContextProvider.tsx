import {
  type FC,
  type PropsWithChildren,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react';
import { ErrorContext } from '../errorHandling/ErrorContext';
import { ErrorTypes } from '../errorHandling/utils';
import {
  useOwnedInstancesQuery,
  type UpdatedOwnedInstancesSubscription,
} from '../generated-types';
import { updatedOwnedInstances } from '../graphql-components/subscription';
import { TenantContext } from './TenantContext';
import { AuthContext } from './AuthContext';
import { OwnedInstancesContext } from './OwnedInstancesContext';
import type { Instance } from '../utils';
import {
  makeGuiInstance,
  getSubObjTypeK8s,
  notifyStatus,
  SubObjType,
} from '../utilsLogic';
import { matchK8sObject, replaceK8sObject } from '../k8sUtils';

const OwnedInstancesContextProvider: FC<PropsWithChildren> = props => {
  const { children } = props;
  const { userId } = useContext(AuthContext);
  const { data: tenantData, notify: notifier } = useContext(TenantContext);
  const { makeErrorCatcher, apolloErrorCatcher, errorsQueue } =
    useContext(ErrorContext);

  const [instances, setInstances] = useState<Instance[]>([]);

  const tenantNs = tenantData?.tenant?.status?.personalNamespace?.name;

  const {
    data,
    loading,
    error,
    refetch: refetchQuery,
    subscribeToMore,
  } = useOwnedInstancesQuery({
    variables: { tenantNamespace: tenantNs || '' },
    skip: !tenantNs,
    onError: apolloErrorCatcher,
    fetchPolicy: 'cache-and-network',
    onCompleted: data => {
      // Convert GraphQL instances to GUI instances
      const guiInstances =
        data?.instanceList?.instances
          ?.map(i => makeGuiInstance(i, userId ?? ''))
          .filter((i): i is Instance => i !== null) ?? [];
      setInstances(guiInstances);
    },
  });

  // Keep track of raw instances for quota calculations
  const rawInstances = useMemo(() => {
    return (
      data?.instanceList?.instances?.filter(
        (i): i is NonNullable<typeof i> => i != null,
      ) ?? []
    );
  }, [data]);

  // Set up subscription for real-time updates
  useEffect(() => {
    if (!tenantNs || loading || error || errorsQueue.length) return;

    const unsubscribe = subscribeToMore<UpdatedOwnedInstancesSubscription>({
      onError: makeErrorCatcher(ErrorTypes.GenericError),
      document: updatedOwnedInstances,
      variables: { tenantNamespace: tenantNs },
      updateQuery: (prev, { subscriptionData }) => {
        const { data } = subscriptionData;

        if (!data?.updateInstance?.instance) return prev;

        const { instance, updateType } = data.updateInstance;

        // Convert to GUI instance for state updates
        const guiInstance = makeGuiInstance(instance, userId ?? '');

        if (!guiInstance) return prev;

        // Update the raw GraphQL data
        const newData = { ...prev };
        if (!newData.instanceList) {
          newData.instanceList = {
            __typename: 'ItPolitoCrownlabsV1alpha2InstanceList',
            instances: [],
          };
        }

        let instances = [...(newData.instanceList.instances || [])];
        const found = instances.find(matchK8sObject(instance, false));
        const objType = getSubObjTypeK8s(found, instance, updateType);
        let notify = false;

        // Handle different update types
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
          case SubObjType.PublicExposureChange:
            instances = instances.map(replaceK8sObject(instance));
            notify = false;
            break;
          case SubObjType.Drop:
            notify = false;
            break;
          default:
            break;
        }

        // Send notification if needed
        if (notify) {
          notifyStatus(instance.status?.phase, instance, updateType, notifier);
        }

        // Update GUI instances state based on objType
        if (objType !== SubObjType.Drop) {
          setInstances(prevInstances => {
            const index = prevInstances.findIndex(
              i => i.name === guiInstance.name && i.id === guiInstance.id,
            );

            if (objType === SubObjType.Deletion) {
              if (index !== -1) {
                return prevInstances.filter((_, i) => i !== index);
              }
              return prevInstances;
            }

            if (index !== -1) {
              // Update existing instance
              const newInstances = [...prevInstances];
              newInstances[index] = guiInstance;
              return newInstances;
            } else {
              // Add new instance
              return [...prevInstances, guiInstance];
            }
          });
        }

        newData.instanceList = {
          ...newData.instanceList,
          instances,
        };

        return newData;
      },
    });

    return unsubscribe;
  }, [
    tenantNs,
    loading,
    error,
    errorsQueue.length,
    subscribeToMore,
    makeErrorCatcher,
    userId,
    notifier,
  ]);

  const refetch = useCallback(async () => {
    if (!tenantNs) return;
    try {
      await refetchQuery();
    } catch (err) {
      console.error('Error refetching owned instances:', err);
    }
  }, [tenantNs, refetchQuery]);

  const contextValue = useMemo(
    () => ({
      data,
      rawInstances,
      instances,
      loading,
      error: error ? new Error(error.message) : undefined,
      refetch,
    }),
    [data, rawInstances, instances, loading, error, refetch],
  );

  return (
    <OwnedInstancesContext.Provider value={contextValue}>
      {children}
    </OwnedInstancesContext.Provider>
  );
};

export default OwnedInstancesContextProvider;
