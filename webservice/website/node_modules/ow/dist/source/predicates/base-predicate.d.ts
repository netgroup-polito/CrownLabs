import { Main } from '..';
/**
@hidden
*/
export declare const testSymbol: unique symbol;
/**
@hidden
*/
export declare const isPredicate: (value: any) => value is BasePredicate<unknown>;
/**
@hidden
*/
export interface BasePredicate<T = unknown> {
    [testSymbol](value: T, main: Main, label: string | Function): void;
}
