import { Informer, ListPromise, ObjectCallback } from './informer';
import { KubernetesObject } from './types';
import { Watch } from './watch';
export interface ObjectCache<T> {
    get(name: string, namespace?: string): T | undefined;
    list(namespace?: string): ReadonlyArray<T>;
}
export declare class ListWatch<T extends KubernetesObject> implements ObjectCache<T>, Informer<T> {
    private readonly path;
    private readonly watch;
    private readonly listFn;
    private objects;
    private readonly indexCache;
    private readonly callbackCache;
    constructor(path: string, watch: Watch, listFn: ListPromise<T>, autoStart?: boolean);
    start(): Promise<void>;
    on(verb: string, cb: ObjectCallback<T>): void;
    off(verb: string, cb: ObjectCallback<T>): void;
    get(name: string, namespace?: string): T | undefined;
    list(namespace?: string | undefined): ReadonlyArray<T>;
    private doneHandler;
    private addOrUpdateItems;
    private indexObj;
    private watchHandler;
}
export declare function deleteItems<T extends KubernetesObject>(oldObjects: T[], newObjects: T[], deleteCallback?: Array<ObjectCallback<T>>): T[];
export declare function addOrUpdateObject<T extends KubernetesObject>(objects: T[], obj: T, addCallback?: Array<ObjectCallback<T>>, updateCallback?: Array<ObjectCallback<T>>): void;
export declare function deleteObject<T extends KubernetesObject>(objects: T[], obj: T, deleteCallback?: Array<ObjectCallback<T>>): void;
