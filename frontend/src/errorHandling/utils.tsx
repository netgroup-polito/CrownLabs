import type { ApolloError } from '@apollo/client';
import type { ErrorContext } from 'react-oidc-context';

export enum ErrorTypes {
  ApolloError,
  AuthError,
  RenderError,
  GenericError,
}

export type ApolloErrorCatcher = {
  onError: (err: ApolloError) => void;
};

export type SupportedError = ApolloError | ErrorContext | Error;

export type EnrichedError = SupportedError & {
  entity?: string;
};

export class CustomError {
  private type: ErrorTypes;
  private error: SupportedError;
  constructor(type: ErrorTypes, error: SupportedError) {
    this.type = type;
    this.error = error;
  }
  getType = () => this.type;
  getError = <T extends SupportedError = SupportedError>(): T =>
    this.error as T;
  getErrorMessage = (): string => {
    let err;
    switch (this.type) {
      case ErrorTypes.RenderError:
        err = this.error as Error;
        return err.message;
      case ErrorTypes.AuthError:
        err = this.error as ErrorContext;
        return err.message;
      case ErrorTypes.ApolloError:
        err = this.error as ApolloError;
        return err.message;
      default:
        err = this.error as Error;
        return err.message;
    }
  };
}

export const hasRenderingError = (errorsQueue: CustomError[]) =>
  errorsQueue.map(e => e.getType()).includes(ErrorTypes.RenderError);
