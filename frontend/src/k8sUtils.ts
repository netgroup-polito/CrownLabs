export interface K8sObject {
  metadata?: {
    name?: string | null;
    namespace?: string | null;
  } | null;
}

export interface prettyNamedK8sObject {
  spec?: {
    prettyName?: string | null;
  } | null;
}

/**
 * Compare two k8s objects.
 */
export function compareK8sObjects(
  a?: K8sObject | null,
  b?: K8sObject | null,
): boolean {
  return (
    a?.metadata?.name === b?.metadata?.name &&
    a?.metadata?.namespace === b?.metadata?.namespace
  );
}

/**
 * Compare two k8s extends objects.
 */
export function comparePrettyName(
  a?: prettyNamedK8sObject | null,
  b?: prettyNamedK8sObject | null,
): boolean {
  return a?.spec?.prettyName === b?.spec?.prettyName;
}

/**
 * Create a lambda for comparing the given object with the one passed to the lambda
 * @param obj object against which compare the ones passed to the lambda
 * @param inverseMatch if true, the lambda will return true if the objects are not equal
 */
export function matchK8sObject(
  obj: K8sObject,
  inverseMatch?: boolean,
): (i?: K8sObject | null) => boolean {
  return (o?: K8sObject | null) => {
    const match = compareK8sObjects(obj, o);
    return inverseMatch ? !match : match;
  };
}

/**
 * Create a lambda which returns the given object if metadata matches the given object name and namespace
 * @param obj object against which compare the ones passed to the lambda
 */
export function replaceK8sObject<T extends K8sObject>(
  obj: T,
): (o: T | null) => T | null {
  return (o: T | null) => {
    return compareK8sObjects(obj, o) ? obj : o;
  };
}
