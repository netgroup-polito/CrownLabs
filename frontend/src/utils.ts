import type { Dispatch, ReactNode, SetStateAction } from 'react';
import type { EnvironmentType, Phase, Phase3 } from './generated-types';
import { Role } from './generated-types';
export type someKeysOf<T> = { [key in keyof T]?: T[key] };
export enum WorkspaceRole {
  user = Role.User,
  manager = Role.Manager,
  candidate = Role.Candidate,
}
export type BadgeSize = 'small' | 'middle' | 'large';
export type User = { tenantId: string; tenantNamespace: string };
export type BoxHeaderSize = 'small' | 'middle' | 'large';
export type Workspace = {
  name: string;
  namespace: string;
  prettyName: string;
  role: WorkspaceRole;
  templates?: Array<Template>;
  waitingTenants?: number;
};
export type Resources = {
  cpu: number;
  disk: string;
  memory: string;
};
export type Template = {
  id: string;
  name: string;
  gui: boolean;
  persistent: boolean;
  nodeSelector?: JSON;
  resources: Resources;
  instances: Array<Instance>;
  workspaceName: string;
  workspaceNamespace: string;
};

export type Instance = {
  id: string;
  gui: boolean;
  templateId: string;
  templateName: string;
  templatePrettyName: string;
  persistent: boolean;
  tenantId: string;
  tenantDisplayName: string;
  tenantNamespace: string;
  environmentType: EnvironmentType;
  name: string;
  prettyName: string;
  ip: string;
  status: Phase;
  url: string | null;
  timeStamp: string;
  workspaceName: string;
  running: boolean;
  nodeSelector?: JSON;
  nodeName?: string;
  myDriveUrl: string;
};

export type SharedVolume = {
  id: string;
  name: string;
  prettyName: string;
  size: string;
  status: Phase3;
  timeStamp: string;
  namespace: string;
};

export enum LinkPosition {
  MenuButton,
  NavbarButton,
  Hidden,
}

export enum WorkspacesAvailableAction {
  None,
  Join,
  AskToJoin,
  Waiting,
}

export type WorkspacesAvailable = {
  name: string;
  prettyName: string;
  role: WorkspaceRole | null;
  action?: WorkspacesAvailableAction;
};

export const generateAvatarUrl = (style: string, seed: string) => {
  return `https://api.dicebear.com/8.x/${style}/svg?seed=${stringHash(seed)}`;
};

export const stringHash = (s: string) => {
  return s.split('').reduce((a, b) => {
    a = (a << 5) - a + b.charCodeAt(0);
    return a & a;
  }, 0);
};

export type RouteData = {
  name: string;
  path: string;
  navbarMenuIcon?: ReactNode;
};

export type RouteDescriptor = {
  route: RouteData;
  content?: ReactNode;
  linkPosition: LinkPosition;
};

export function multiStringIncludes(needle: string, ...haystack: string[]) {
  needle = needle.toLowerCase().replace(/\s/g, '');
  const concatenatedString = haystack.join('').toLowerCase().replace(/\s/g, '');

  return concatenatedString.includes(needle);
}

/**
 * Create a callback that can be used to set a list state, by toggling the presence of the list of a given value.
 * @param setList the setter for the list
 * @param create specify if the returned function is used to create a new instance or not
 * @returns a callback which accepts a value and toggles the presence of that value in the list
 */
export function makeListToggler<T>(
  setList: Dispatch<SetStateAction<Array<T>>>,
): (value: T, create: boolean) => void {
  return (value: T, create: boolean) => {
    setList(list =>
      list.includes(value)
        ? create
          ? list
          : list.filter(v => v !== value)
        : [...list, value],
    );
  };
}

export const JSONDeepCopy = <T>(obj: T) =>
  obj && (JSON.parse(JSON.stringify(obj)) as T);

export type WorkspaceEntry = { role: Role; name: string };

export type UserAccountPage = {
  key: string;
  userid: string;
  name: string;
  surname: string;
  email: string;
  currentRole?: string;
  workspaces?: WorkspaceEntry[];
};

export function makeRandomDigits(value: number) {
  return Math.random().toFixed(value).replace('0.', '');
}

export function filterUser(user: UserAccountPage, value: string) {
  return multiStringIncludes(
    value,
    user.name,
    user.surname,
    user.userid,
    user.userid,
  );
}

/**
 * Find the key for a given value of an Enum.
 * @param obj the enumeration
 * @param value the value of the enumeration
 * @returns the (first) key corresponding to the passed value or undefined
 */
export const findKeyByValue = <T, K extends keyof unknown>(
  obj: Record<K, T>,
  value: T,
): K | undefined => (Object.keys(obj) as K[]).find(key => obj[key] === value);

/**
 * Converts a string in k8s Resource.Quantity format to a number in GiB.
 * @param sizeStr the string to convert (e.g. '2048Mi')
 * @returns the number that represents the passed quantity in GiB (e.g. 2)
 */
export const convertToGiB = (sizeStr: string): number => {
  const regexp = /[0-9]+(\.[0-9]+)?/g;
  const match = sizeStr.match(regexp);
  if (!match) {
    throw new Error('Invalid size string');
  }
  const num = parseFloat(match[0]);
  if (sizeStr.toLowerCase().includes('gi')) {
    return num;
  } else if (sizeStr.toLowerCase().includes('mi')) {
    return num / 1024;
  } else if (sizeStr.toLowerCase().includes('ki')) {
    return num / (1024 * 1024);
  } else if (sizeStr.toLowerCase().includes('g')) {
    return num * 0.9313225746154785;
  } else if (sizeStr.toLowerCase().includes('m')) {
    return (num / 1024) * 0.9313225746154785;
  } else if (sizeStr.toLowerCase().includes('k')) {
    return (num / (1024 * 1024)) * 0.9313225746154785;
  } else {
    throw new Error('Unsupported size unit');
  }
};

/**
 * Converts a string in k8s Resource.Quantity format to a number in GB.
 * @param sizeStr the string to convert (e.g. '2000M')
 * @returns the number that represents the passed quantity in GB (e.g. 2)
 */
export const convertToGB = (sizeStr: string): number => {
  return convertToGiB(sizeStr) * 1.073741824;
};

/**
 * Approximates a number to the n-th decimal place.
 * @param value the number to approximate
 * @param n the number of decimal places to keep
 * @returns the approximated number
 */
export const approximate = (value: number, n: number): number => {
  const factor = Math.pow(10, n);
  return Math.round(value * factor) / factor;
};

export const camelize = (str: string) =>
  str
    .replace(/(?:^\w|[A-Z]|\b\w|\s+)/g, (match, index) =>
      +match === 0
        ? ''
        : index === 0
          ? match.toLowerCase()
          : match.toUpperCase(),
    )
    .replace(/-/g, '');

export const cleanupLabels = (s?: string) =>
  camelize(
    s?.replace('crownlabs.polito.it/', '').replace('crownlabsPolitoIt', '') ||
      '',
  );

export function enumKeyFromVal<T extends Record<string, string | number>>(
  enumObj: T,
  value: string | number,
): keyof T | undefined {
  return (Object.keys(enumObj) as (keyof T)[]).find(
    key => enumObj[key] === value,
  );
}
