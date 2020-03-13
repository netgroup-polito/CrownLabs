import { Predicate, PredicateOptions } from './predicate';
export declare class NumberPredicate extends Predicate<number> {
    /**
    @hidden
    */
    constructor(options?: PredicateOptions);
    /**
    Test a number to be in a specified range.

    @param start - Start of the range.
    @param end - End of the range.
    */
    inRange(start: number, end: number): this;
    /**
    Test a number to be greater than the provided value.

    @param number - Minimum value.
    */
    greaterThan(number: number): this;
    /**
    Test a number to be greater than or equal to the provided value.

    @param number - Minimum value.
    */
    greaterThanOrEqual(number: number): this;
    /**
    Test a number to be less than the provided value.

    @param number - Maximum value.
    */
    lessThan(number: number): this;
    /**
    Test a number to be less than or equal to the provided value.

    @param number - Minimum value.
    */
    lessThanOrEqual(number: number): this;
    /**
    Test a number to be equal to a specified number.

    @param expected - Expected value to match.
    */
    equal(expected: number): this;
    /**
    Test if a number is an element of the provided list.

    @param list - List of possible values.
    */
    oneOf(list: readonly number[]): this;
    /**
    Test a number to be an integer.
    */
    readonly integer: this;
    /**
    Test a number to be finite.
    */
    readonly finite: this;
    /**
    Test a number to be infinite.
    */
    readonly infinite: this;
    /**
    Test a number to be positive.
    */
    readonly positive: this;
    /**
    Test a number to be negative.
    */
    readonly negative: this;
    /**
    Test a number to be an integer or infinite.
    */
    readonly integerOrInfinite: this;
    /**
    Test a number to be in a valid range for a 8-bit unsigned integer.
    */
    readonly uint8: this;
    /**
    Test a number to be in a valid range for a 16-bit unsigned integer.
    */
    readonly uint16: this;
    /**
    Test a number to be in a valid range for a 32-bit unsigned integer.
    */
    readonly uint32: this;
    /**
    Test a number to be in a valid range for a 8-bit signed integer.
    */
    readonly int8: this;
    /**
    Test a number to be in a valid range for a 16-bit signed integer.
    */
    readonly int16: this;
    /**
    Test a number to be in a valid range for a 32-bit signed integer.
    */
    readonly int32: this;
}
