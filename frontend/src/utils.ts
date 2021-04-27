export type someKeysOf<T> = { [key in keyof T]?: T[key] };
export type WorkspaceRole = 'user' | 'manager';
export type BadgeSize = 'small' | 'middle' | 'large';
export type BoxHeaderSize = 'small' | 'middle' | 'large';
