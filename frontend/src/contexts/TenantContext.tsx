import { ApolloError } from '@apollo/client';
import {
  createContext,
  FC,
  PropsWithChildren,
  useContext,
  useEffect,
  useState,
} from 'react';
import { AuthContext } from './AuthContext';
import { ErrorContext } from '../errorHandling/ErrorContext';
import { ErrorTypes } from '../errorHandling/utils';
import {
  TenantQuery,
  UpdatedTenantSubscription,
  useApplyTenantMutation,
  useTenantQuery,
} from '../generated-types';
import { JSONDeepCopy } from '../utils';
import { workspaceGetName } from '../utilsLogic';
import { updatedTenant } from '../graphql-components/subscription';
import { getTenantPatchJson } from '../graphql-components/utils';

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

  const { makeErrorCatcher, apolloErrorCatcher, errorsQueue } =
    useContext(ErrorContext);

  const { loading, error, subscribeToMore } = useTenantQuery({
    variables: { tenantId: userId ?? '' },
    onCompleted: d => {
      d!.tenant!.spec?.workspaces!.sort((a, b) =>
        workspaceGetName(a!).localeCompare(workspaceGetName(b!))
      );
      setData(JSONDeepCopy(d));
    },
    fetchPolicy: 'network-only',
    onError: apolloErrorCatcher,
  });

  const [applyTenantMutation] = useApplyTenantMutation({
    onError: apolloErrorCatcher,
  });

  useEffect(() => {
    if (!loading && !error && !errorsQueue.length) {
      const unsubscribe = subscribeToMore({
        onError: makeErrorCatcher(ErrorTypes.GenericError),
        variables: { tenantId: userId ?? '' },
        document: updatedTenant,
        updateQuery: (prev, { subscriptionData: { data } }) => {
          const dataCasted = data as UpdatedTenantSubscription;
          if (!data) return prev;
          setData(dataCasted.updatedTenant as TenantQuery);
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

  const refreshClock = () => setNow(new Date());

  useEffect(() => {
    const timerHandler = setInterval(refreshClock, 60000);
    return () => clearInterval(timerHandler);
  }, []);

  useEffect(() => {
    if (!userId) return;
    applyTenantMutation({
      variables: {
        tenantId: userId,
        patchJson: getTenantPatchJson({
          lastLogin: new Date(),
        }),
        manager: 'frontend-tenant-lastlogin',
      },
      onError: apolloErrorCatcher,
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [userId]);

  const hasSSHKeys = !!data?.tenant?.spec?.publicKeys?.length;
  return (
    <TenantContext.Provider
      value={{ data, loading, error, now, refreshClock, hasSSHKeys }}
    >
      {children}
    </TenantContext.Provider>
  );
};

export default TenantContextProvider;
