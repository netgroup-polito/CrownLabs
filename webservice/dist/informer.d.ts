/// <reference types="node" />
import { KubeConfig } from './config';
import { KubernetesListObject, KubernetesObject } from './types';
import http = require('http');
export declare type ObjectCallback<T extends KubernetesObject> = (obj: T) => void;
export declare type ListCallback<T extends KubernetesObject> = (list: T[], ResourceVersion: string) => void;
export declare type ListPromise<T extends KubernetesObject> = () => Promise<{
    response: http.IncomingMessage;
    body: KubernetesListObject<T>;
}>;
export declare const ADD: string;
export declare const UPDATE: string;
export declare const DELETE: string;
export interface Informer<T> {
    on(verb: string, fn: ObjectCallback<T>): any;
    off(verb: string, fn: ObjectCallback<T>): any;
    start(): Promise<void>;
}
export declare function makeInformer<T>(kubeconfig: KubeConfig, path: string, listPromiseFn: ListPromise<T>): Informer<T>;
