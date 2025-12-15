import type { UpdateType } from '../generated-types';
import { matchK8sObject, replaceK8sObject } from '../k8sUtils';
import { getSubObjTypeK8s, notifyStatus, SubObjType } from '../utilsLogic';
import type { Notifier } from '../contexts/TenantContext';

// Use a more flexible type that matches the GraphQL query result structure
// eslint-disable-next-line @typescript-eslint/no-explicit-any
type InstanceType = any;

interface InstanceUpdateData {
  instance: InstanceType;
  updateType: UpdateType;
}

interface SubscriptionHandlerOptions {
  tenantNamespace?: string;
  notifier?: Notifier;
  shouldNotify?: (namespace: string, tenantNamespace: string) => boolean;
}

/**
 * Generic handler for instance subscription updates
 * Returns updated instances array and whether to notify
 */
export function handleInstanceUpdate(
  prev: { instanceList?: { instances?: InstanceType[]; __typename?: string } },
  instanceUpdate: InstanceUpdateData,
  options: SubscriptionHandlerOptions = {},
): {
  instances: InstanceType[];
  shouldNotify: boolean;
  objType: SubObjType;
} {
  const { instance, updateType } = instanceUpdate;
  const { tenantNamespace, notifier, shouldNotify } = options;

  if (!prev.instanceList?.instances) {
    return {
      instances: [],
      shouldNotify: false,
      objType: SubObjType.Drop,
    };
  }

  let instances = [...prev.instanceList.instances];
  const found = instances.find(matchK8sObject(instance, false));
  const objType = getSubObjTypeK8s(found, instance, updateType);

  let notify = false;
  const namespace = instance.metadata?.namespace;
  const matchNS = tenantNamespace
    ? shouldNotify
      ? shouldNotify(namespace ?? '', tenantNamespace)
      : namespace === tenantNamespace
    : false;

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
  if (notify && matchNS && notifier) {
    notifyStatus(instance.status?.phase, instance, updateType, notifier);
  }

  return {
    instances,
    shouldNotify: notify && matchNS,
    objType,
  };
}

/**
 * Creates the updateQuery function for subscribeToMore
 */
export function createInstanceUpdateQuery(
  options: SubscriptionHandlerOptions = {},
) {
   
  return <
    T extends {
      instanceList?: { instances?: any[]; __typename?: string } | null;
    },
  >(
    prev: T,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    { subscriptionData }: { subscriptionData: any },
  ): T => {
    const data = subscriptionData?.data;

    // Handle both UpdatedOwnedInstancesSubscription and UpdatedInstancesLabelSelectorSubscription
    const instanceUpdate =
      data?.updateInstance || data?.updateInstanceLabelSelector;

    if (!instanceUpdate?.instance || !instanceUpdate?.updateType) return prev;

    const { instances } = handleInstanceUpdate(
      { instanceList: prev.instanceList ?? undefined },
      {
        instance: instanceUpdate.instance,
        updateType: instanceUpdate.updateType,
      },
      options,
    );

    return {
      ...prev,
      instanceList: {
        __typename: prev.instanceList?.__typename,
        instances,
      },
    };
  };
}
