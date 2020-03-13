import { Predicate, PredicateOptions } from './predicate';
export declare class BooleanPredicate extends Predicate<boolean> {
    /**
    @hidden
    */
    constructor(options?: PredicateOptions);
    /**
    Test a boolean to be true.
    */
    readonly true: this;
    /**
    Test a boolean to be false.
    */
    readonly false: this;
}
