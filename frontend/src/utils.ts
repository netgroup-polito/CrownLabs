export type someKeysOf<T> = { [key in keyof T]?: T[key] };
export type WorkspaceRole = 'user' | 'manager';
export type BadgeSize = 'small' | 'middle' | 'large';
export type BoxHeaderSize = 'small' | 'middle' | 'large';
export type Template = {
  id: string;
  name: string;
  gui: boolean;
  instances: Array<Instance>;
};
export type Instance = {
  id: number;
  name: string;
  ip: string;
  status: boolean;
};
