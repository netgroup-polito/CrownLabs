/* eslint-disable react/no-multi-comp */

import { ApolloError } from '@apollo/client';
import {
  Component,
  createContext,
  Dispatch,
  FC,
  PropsWithChildren,
  ReactNode,
  SetStateAction,
  useEffect,
  useState,
} from 'react';
import { ErrorHandler } from './ErrorHandler';
import RenderErrorHandler from './RenderErrorHandler/RenderErrorHandler';
import {
  CustomError,
  ErrorTypes,
  hasRenderingError,
  SupportedError,
} from './utils';

interface IErrorContext {
  errorsQueue: Array<CustomError>;
  makeErrorCatcher: <T extends SupportedError>(
    errType: ErrorTypes
  ) => (err: T) => void;
  apolloErrorCatcher: (err: ApolloError) => void;
  getNextError: () => void;
  flushRenderError: () => void;
  execLogin: boolean;
  setExecLogin: Dispatch<SetStateAction<boolean>>;
}

interface PropsErrorBoundary {
  children: ReactNode;
  makeErrorCatcher: <T extends SupportedError>(
    errType: ErrorTypes
  ) => (err: T) => void;
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

interface StateErrorBoundary {
  hasError: boolean;
}

class ErrorBoundary extends Component<PropsErrorBoundary, StateErrorBoundary> {
  constructor(props: PropsErrorBoundary) {
    super(props);
    this.state = { hasError: false };
  }
  static getDerivedStateFromError(error: Error) {
    return { hasError: true };
  }
  private catchError = this.props.makeErrorCatcher(ErrorTypes.RenderError);
  componentDidCatch(error: Error) {
    this.catchError(error);
  }
  render() {
    return this.props.children;
  }
}

const ErrorContextProvider: FC<PropsWithChildren<{}>> = props => {
  const { children } = props;

  const [errorsQueue, setErrorsQueue] = useState<Array<CustomError>>([]);

  const [execLogin, setExecLogin] = useState(true);
  useEffect(() => {
    if (
      errorsQueue.find(e =>
        (e.getErrorMessage() || '').includes('Unauthorized')
      )
    )
      setExecLogin(true);
  }, [errorsQueue]);

  const dispatchError = (err: CustomError) => {
    setErrorsQueue(old =>
      !old.find(e => e.getErrorMessage() === err.getErrorMessage())
        ? [err, ...old]
        : [...old]
    );
  };

  const getNextError = () => setErrorsQueue(old => old.slice(1));

  const flushRenderError = () => {
    setErrorsQueue(old =>
      old.filter(e => e.getType() !== ErrorTypes.RenderError)
    );
    setExecLogin(true);
  };

  const makeErrorCatcher = <T extends SupportedError>(
    errorType: ErrorTypes
  ) => {
    return (err: T) => dispatchError(new CustomError(errorType, err));
  };

  const apolloErrorCatcher = makeErrorCatcher(ErrorTypes.ApolloError);

  const filteredErrorQueue = errorsQueue.filter(
    e => e.getType() !== ErrorTypes.RenderError
  );

  const renderErrorQueue = errorsQueue.filter(
    e => e.getType() === ErrorTypes.RenderError
  );

  return (
    <ErrorContext.Provider
      value={{
        errorsQueue,
        makeErrorCatcher,
        apolloErrorCatcher,
        getNextError,
        setExecLogin,
        execLogin,
        flushRenderError,
      }}
    >
      <ErrorBoundary makeErrorCatcher={makeErrorCatcher}>
        <ErrorHandler
          errorsQueue={filteredErrorQueue}
          show={filteredErrorQueue.length > 0}
          dismiss={getNextError}
        />
        {!hasRenderingError(errorsQueue) ? (
          children
        ) : (
          <RenderErrorHandler errors={renderErrorQueue} />
        )}
      </ErrorBoundary>
    </ErrorContext.Provider>
  );
};

export default ErrorContextProvider;
