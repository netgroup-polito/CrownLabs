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
import {
  TenantQuery,
  useSshKeysQuery,
  useTenantQuery,
} from '../../generated-types';
import { JSON_StringifyAndParse } from '../../utils';
import { workspaceGetName } from '../../utilsLogic';
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

  const { loading, error, subscribeToMore } = useTenantQuery({
    variables: { tenantId: userId ?? '' },
    onCompleted: d => {
      d!.tenant!.spec?.workspaces!.sort((a, b) =>
        workspaceGetName(a).localeCompare(workspaceGetName(b))
      );
      setData(JSON_StringifyAndParse(d));
    },
    fetchPolicy: 'network-only',
  });

  useEffect(() => {
    if (loading) return;
    subscribeToMore({
      variables: { tenantId: userId ?? '' },
      document: updatedTenant,
      updateQuery: (prev, { subscriptionData: { data } }) => {
        if (!data) return prev;
        data!.tenant!.spec?.workspaces!.sort((a, b) =>
          workspaceGetName(a).localeCompare(workspaceGetName(b))
        );
        setData(JSON_StringifyAndParse(data));
        return data;
      },
    });
  }, [subscribeToMore, loading, userId]);

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
