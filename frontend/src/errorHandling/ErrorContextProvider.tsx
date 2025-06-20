import {
  Component,
  useCallback,
  useEffect,
  useState,
  type FC,
  type PropsWithChildren,
  type ReactNode,
} from 'react';
import {
  CustomError,
  ErrorTypes,
  hasRenderingError,
  type SupportedError,
} from './utils';
import { ErrorContext } from './ErrorContext';
import { ErrorHandler } from './ErrorHandler';
import RenderErrorHandler from './RenderErrorHandler';
import type { ApolloError } from '@apollo/client';
import type { GraphQLFormattedError as GQLFErr } from 'graphql';

interface PropsErrorBoundary {
  children: ReactNode;
  makeErrorCatcher: <T extends SupportedError>(
    errType: ErrorTypes,
  ) => (err: T) => void;
}

interface StateErrorBoundary {
  hasError: boolean;
}

class ErrorBoundary extends Component<PropsErrorBoundary, StateErrorBoundary> {
  constructor(props: PropsErrorBoundary) {
    super(props);
    this.state = { hasError: false };
  }
  static getDerivedStateFromError(_error: Error) {
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

const ErrorContextProvider: FC<PropsWithChildren> = props => {
  const { children } = props;

  const [errorsQueue, setErrorsQueue] = useState<Array<CustomError>>([]);

  const [execLogin, setExecLogin] = useState(false);
  useEffect(() => {
    if (
      errorsQueue.find(
        e =>
          (e.getError<ApolloError>().cause as GQLFErr)?.extensions
            ?.statusCode == 401 ||
          (e.getErrorMessage() || '').includes('Unauthorized'),
      )
    )
      setExecLogin(true);
  }, [errorsQueue]);

  const dispatchError = (err: CustomError) => {
    console.trace(err);
    setErrorsQueue(old =>
      !old.find(e => e.getErrorMessage() === err.getErrorMessage())
        ? [err, ...old]
        : [...old],
    );
  };

  const getNextError = () => setErrorsQueue(old => old.slice(1));

  const flushRenderError = () => {
    setErrorsQueue(old =>
      old.filter(e => e.getType() !== ErrorTypes.RenderError),
    );
    setExecLogin(true);
  };

  const makeErrorCatcher = useCallback(
    <T extends SupportedError>(errorType: ErrorTypes) => {
      return (err: T) => dispatchError(new CustomError(errorType, err));
    },
    [],
  );

  const apolloErrorCatcher = makeErrorCatcher(ErrorTypes.ApolloError);

  const filteredErrorQueue = errorsQueue.filter(
    e => e.getType() !== ErrorTypes.RenderError,
  );

  const renderErrorQueue = errorsQueue.filter(
    e => e.getType() === ErrorTypes.RenderError,
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
