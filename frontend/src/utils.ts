import { Dispatch, ReactNode, SetStateAction } from 'react';
import { EnvironmentType, Phase } from './generated-types';
import { Role } from './generated-types';
export type someKeysOf<T> = { [key in keyof T]?: T[key] };
export enum WorkspaceRole {
  user = 'user',
  manager = 'manager',
  candidate = 'candidate',
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
};

export enum LinkPosition {
  MenuButton,
  NavbarButton,
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
  var concatenatedString = haystack.join('').toLowerCase().replace(/\s/g, '');

  return concatenatedString.includes(needle);
}

/**
 * Create a callback that can be used to set a list state, by toggling the presence of the list of a given value.
 * @param setList the setter for the list
 * @param create specify if the returned function is used to create a new instance or not
 * @returns a callback which accepts a value and toggles the presence of that value in the list
 */
export function makeListToggler<T>(
  setList: Dispatch<SetStateAction<Array<T>>>
): (value: T, create: boolean) => void {
  return (value: T, create: boolean) => {
    setList(list =>
      list.includes(value)
        ? create
          ? list
          : list.filter(v => v !== value)
        : [...list, value]
    );
  };
}

export const JSONDeepCopy = <T>(obj: T) => JSON.parse(JSON.stringify(obj)) as T;

export type UserAccountPage = {
  key: string;
  userid: string;
  name: string;
  surname: string;
  email: string;
  currentRole?: string;
  workspaces?: { role: Role; name: string }[];
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
    user.userid
  );
}
