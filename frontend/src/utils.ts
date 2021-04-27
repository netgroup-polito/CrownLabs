export type someKeysOf<T> = { [key in keyof T]?: T[key] };
