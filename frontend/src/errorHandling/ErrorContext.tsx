import type { ApolloError } from '@apollo/client';
import { createContext, type Dispatch, type SetStateAction } from 'react';
import type { CustomError, ErrorTypes } from './utils';
import { type SupportedError } from './utils';

interface IErrorContext {
  errorsQueue: Array<CustomError>;
  makeErrorCatcher: <T extends SupportedError>(
    errType: ErrorTypes,
  ) => (err: T) => void;
  apolloErrorCatcher: (err: ApolloError) => void;
  getNextError: () => void;
  flushRenderError: () => void;
  execLogin: boolean;
  setExecLogin: Dispatch<SetStateAction<boolean>>;
}

export const ErrorContext = createContext<IErrorContext>({
  errorsQueue: [],
  makeErrorCatcher: () => () => null,
  apolloErrorCatcher: () => null,
  getNextError: () => null,
  flushRenderError: () => null,
  execLogin: true,
  setExecLogin: () => null,
});
