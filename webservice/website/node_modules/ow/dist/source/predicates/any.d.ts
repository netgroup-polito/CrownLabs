import { BasePredicate, testSymbol } from './base-predicate';
import { Main } from '..';
import { PredicateOptions } from './predicate';
/**
@hidden
*/
export declare class AnyPredicate<T = unknown> implements BasePredicate<T> {
    private readonly predicates;
    private readonly options;
    constructor(predicates: BasePredicate[], options?: PredicateOptions);
    [testSymbol](value: T, main: Main, label: string | Function): void;
}
