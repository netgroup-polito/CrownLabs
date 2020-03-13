import { Main } from '..';
import { BasePredicate, testSymbol } from './base-predicate';
/**
@hidden
*/
export interface Validator<T> {
    message(value: T, label?: string, result?: any): string;
    validator(value: T): unknown;
}
/**
@hidden
*/
export interface PredicateOptions {
    optional?: boolean;
}
/**
@hidden
*/
export interface Context<T = unknown> extends PredicateOptions {
    validators: Validator<T>[];
}
/**
@hidden
*/
export declare const validatorSymbol: unique symbol;
export declare type CustomValidator<T> = (value: T) => {
    /**
    Should be `true` if the validation is correct.
    */
    validator: boolean;
    /**
    The error message which should be shown if the `validator` is `false`. Or a error function which returns the error message and accepts the label as first argument.
    */
    message: string | ((label: string) => string);
};
/**
@hidden
*/
export declare class Predicate<T = unknown> implements BasePredicate<T> {
    private readonly type;
    private readonly options;
    private readonly context;
    constructor(type: string, options?: PredicateOptions);
    /**
    @hidden
    */
    [testSymbol](value: T, main: Main, label: string | Function): void;
    /**
    @hidden
    */
    readonly [validatorSymbol]: Validator<T>[];
    /**
    Invert the following validators.
    */
    readonly not: this;
    /**
    Test if the value matches a custom validation function. The validation function should return an object containing a `validator` and `message`. If the `validator` is `false`, the validation fails and the `message` will be used as error message. If the `message` is a function, the function is invoked with the `label` as argument to let you further customize the error message.

    @param customValidator - Custom validation function.
    */
    validate(customValidator: CustomValidator<T>): this;
    /**
    Test if the value matches a custom validation function. The validation function should return `true` if the value passes the function. If the function either returns `false` or a string, the function fails and the string will be used as error message.

    @param validator - Validation function.
    */
    is(validator: (value: T) => boolean | string): this;
    /**
    Register a new validator.

    @param validator - Validator to register.
    */
    protected addValidator(validator: Validator<T>): this;
}
