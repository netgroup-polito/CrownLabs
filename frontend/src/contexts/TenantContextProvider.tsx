import {
  type FC,
  type PropsWithChildren,
  useCallback,
  useContext,
  useEffect,
  useState,
} from 'react';
import { useAuth } from 'react-oidc-context';
import { ErrorContext } from '../errorHandling/ErrorContext';
import { ErrorTypes } from '../errorHandling/utils';
import {
  type TenantQuery,
  type UpdatedTenantSubscription,
  useApplyTenantMutation,
  useTenantQuery,
} from '../generated-types';
import { JSONDeepCopy } from '../utils';
import { workspaceGetName } from '../utilsLogic';
import { updatedTenant } from '../graphql-components/subscription';
import { getTenantPatchJson } from '../graphql-components/utils';
import { TenantContext } from './TenantContext';
import { AuthContext } from './AuthContext';

const TenantContextProvider: FC<PropsWithChildren> = props => {
  const { userId } = useContext(AuthContext);
  const { children } = props;

  const [now, setNow] = useState(new Date());
  const [data, setData] = useState<TenantQuery>();

  const auth = useAuth();

  const { makeErrorCatcher, apolloErrorCatcher, errorsQueue } =
    useContext(ErrorContext);

  const { loading, error, subscribeToMore } = useTenantQuery({
    skip: !userId,
    variables: { tenantId: userId || '' },
    onCompleted: data => {
      const d = JSONDeepCopy(data);
      d?.tenant?.spec?.workspaces?.sort((a, b) =>
        workspaceGetName(a).localeCompare(workspaceGetName(b)),
      );
      setData(d);
    },
    fetchPolicy: 'network-only',
    onError: apolloErrorCatcher,
  });

  const [applyTenantMutation] = useApplyTenantMutation({
    onError: apolloErrorCatcher,
  });

  useEffect(() => {
    if (userId && !loading && !error && !errorsQueue.length) {
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

  const patchTenantLastLogin = useCallback((tenantId: string) => {
    applyTenantMutation({
      variables: {
        tenantId,
        patchJson: getTenantPatchJson({
          lastLogin: new Date(),
        }),
        manager: 'frontend-tenant-lastlogin',
      },
      onError: apolloErrorCatcher,
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (!data?.tenant?.metadata?.name || !userId) return;
    patchTenantLastLogin(userId);
  }, [userId, data?.tenant?.metadata?.name, patchTenantLastLogin]);

  const tSpec = data?.tenant?.spec;
  const displayName = tSpec
    ? `${tSpec.firstName} ${tSpec.lastName}`
    : auth.user?.profile?.name || '';

  const hasSSHKeys = !!tSpec?.publicKeys?.length;
  return (
    <TenantContext.Provider
      value={{
        data,
        loading,
        error,
        now,
        refreshClock,
        hasSSHKeys,
        displayName,
      }}
    >
      {children}
    </TenantContext.Provider>
  );
};

export default TenantContextProvider;
