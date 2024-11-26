import { ApolloError } from '@apollo/client';
import { KeycloakError } from 'keycloak-js';

export enum ErrorTypes {
  ApolloError,
  KeycloakError,
  RenderError,
  GenericError,
}

export type ApolloErrorCatcher = {
  onError: (err: ApolloError) => void;
};

export type SupportedError = ApolloError | KeycloakError | Error;

export class CustomError {
  private type: ErrorTypes;
  private error: ApolloError | KeycloakError | Error;
  constructor(type: ErrorTypes, error: ApolloError | KeycloakError | Error) {
    this.type = type;
    this.error = error;
  }
  getType = () => this.type;
  getError = () => this.error;
  getErrorMessage = (): string => {
    let err;
    switch (this.type) {
      case ErrorTypes.RenderError:
        err = this.error as Error;
        return err.message;
      case ErrorTypes.KeycloakError:
        err = this.error as KeycloakError;
        return err.error;
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
