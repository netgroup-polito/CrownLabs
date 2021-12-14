import { ApolloError } from 'apollo-client';
import {
  createContext,
  FC,
  PropsWithChildren,
  useContext,
  useEffect,
  useState,
} from 'react';
import { AuthContext } from '../../contexts/AuthContext';
import { ErrorContext } from '../../errorHandling/ErrorContext';
import { ErrorTypes } from '../../errorHandling/utils';
import {
  TenantQuery,
  useSshKeysQuery,
  useTenantQuery,
} from '../../generated-types';
import { updatedTenant } from '../subscription';

interface ITenantContext {
  data?: TenantQuery;
  loading?: boolean;
  error?: ApolloError;
  hasSSHKeys: boolean;
  now: Date;
  refreshClock: () => void;
}

export const TenantContext = createContext<ITenantContext>({
  data: undefined,
  loading: undefined,
  error: undefined,
  hasSSHKeys: false,
  now: new Date(),
  refreshClock: () => null,
});

const TenantContextProvider: FC<PropsWithChildren<{}>> = props => {
  const { userId } = useContext(AuthContext);
  const { children } = props;

  const [now, setNow] = useState(new Date());
  const [data, setData] = useState<TenantQuery>();

  const { data: sshKeysResult } = useSshKeysQuery({
    variables: { tenantId: userId ?? '' },
    notifyOnNetworkStatusChange: true,
    fetchPolicy: 'network-only',
  });
  const { makeErrorCatcher, apolloErrorCatcher, errorsQueue } =
    useContext(ErrorContext);

  const { loading, error, subscribeToMore } = useTenantQuery({
    variables: { tenantId: userId ?? '' },
    onCompleted: setData,
    fetchPolicy: 'network-only',
    onError: apolloErrorCatcher,
  });

  useEffect(() => {
    if (!loading && !error && !errorsQueue.length) {
      const unsubscribe = subscribeToMore({
        onError: makeErrorCatcher(ErrorTypes.GenericError),
        variables: { tenantId: userId ?? '' },
        document: updatedTenant,
        updateQuery: (prev, { subscriptionData: { data } }) => {
          if (!data) return prev;
          setData(data);
          return data;
        },
      });
      return unsubscribe;
    }
  }, [
    subscribeToMore,
    loading,
    userId,
    errorsQueue.length,
    error,
    apolloErrorCatcher,
    makeErrorCatcher,
  ]);

  useEffect(() => {
    const timerHandler = setInterval(() => setNow(new Date()), 60000);
    return () => clearInterval(timerHandler);
  }, []);

  const refreshClock = () => setNow(new Date());

  const hasSSHKeys = !!sshKeysResult?.tenant?.spec?.publicKeys?.length;

  return (
    <TenantContext.Provider
      value={{ data, loading, error, hasSSHKeys, now, refreshClock }}
    >
      {children}
    </TenantContext.Provider>
  );
};

export default TenantContextProvider;
