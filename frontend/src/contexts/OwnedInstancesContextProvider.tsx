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
import { type Instance } from '../utils';
import { makeGuiInstance, SubObjType } from '../utilsLogic';
import { useQuotaCalculations } from '../components/workspaces/QuotaDisplay/useQuotaCalculation';
import { QuotaContext } from './QuotaContext.types';
import type { ApolloError } from '@apollo/client';
import { handleInstanceUpdate } from '../utils/instanceSubscriptionHandler';
import { calculateWorkspaceConsumedQuota } from '../utils/quota';

const OwnedInstancesContextProvider: FC<PropsWithChildren> = props => {
  const { children } = props;
  const { userId } = useContext(AuthContext);
  const { data: tenantData, notify: notifier } = useContext(TenantContext);
  const { makeErrorCatcher, apolloErrorCatcher, errorsQueue } =
    useContext(ErrorContext);

  const [instances, setInstances] = useState<Instance[]>([]);

  const tenantNs = tenantData?.tenant?.status?.personalNamespace?.name;
  const tenant = tenantData?.tenant ?? undefined;

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
        const data = subscriptionData?.data;

        if (!data?.updateInstance?.instance) return prev;

        const { instance, updateType } = data.updateInstance;

        if (!updateType) return prev;

        // Convert to GUI instance for state updates
        const guiInstance = makeGuiInstance(instance, userId ?? '');

        if (!guiInstance) return prev;

        // Use the shared handler for instance updates
        const { instances, objType } = handleInstanceUpdate(
          { instanceList: prev.instanceList ?? undefined },
          { instance, updateType },
          {
            tenantNamespace: tenantNs,
            notifier,
          },
        );

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

        return {
          ...prev,
          instanceList: {
            __typename: prev.instanceList?.__typename,
            instances,
          },
        };
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

  // Calculate quota using raw instances
  const quotaCalculations = useQuotaCalculations(
    rawInstances as Parameters<typeof useQuotaCalculations>[0],
    tenant,
  );

  // Enhanced refresh function for quota
  const refreshQuota = useCallback(async () => {
    if (!tenantNs) return;
    try {
      await refetch();
    } catch (error) {
      console.error('Error refreshing quota data:', error);
      apolloErrorCatcher(error as ApolloError);
    }
  }, [refetch, apolloErrorCatcher, tenantNs]);

  const consumedQuota = useMemo(
    () => calculateWorkspaceConsumedQuota(instances),
    [instances],
  );

  const contextValue = useMemo(
    () => ({
      data,
      rawInstances,
      instances,
      loading,
      error: error ? new Error(error.message) : undefined,
      refetch,
      consumedQuota,
    }),
    [data, rawInstances, instances, loading, error, refetch],
  );

  const quotaContextValue = useMemo(
    () => ({
      refreshQuota,
      consumedQuota: quotaCalculations.consumedQuota,
      workspaceQuota: quotaCalculations.workspaceQuota,
      availableQuota: quotaCalculations.availableQuota,
    }),
    [refreshQuota, quotaCalculations],
  );

  return (
    <OwnedInstancesContext.Provider value={contextValue}>
      <QuotaContext.Provider value={quotaContextValue}>
        {children}
      </QuotaContext.Provider>
    </OwnedInstancesContext.Provider>
  );
};

export default OwnedInstancesContextProvider;
