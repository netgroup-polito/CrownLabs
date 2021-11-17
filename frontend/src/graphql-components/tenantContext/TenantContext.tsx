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
import { TenantQuery, useTenantQuery } from '../../generated-types';
import { updatedTenant } from '../subscription';

interface ITenantContext {
  data?: TenantQuery;
  loading?: boolean;
  error?: ApolloError;
}

export const TenantContext = createContext<ITenantContext>({
  data: undefined,
  loading: undefined,
  error: undefined,
});

const TenantContextProvider: FC<PropsWithChildren<{}>> = props => {
  const { userId } = useContext(AuthContext);
  const { children } = props;

  const [data, setData] = useState<TenantQuery>();

  const { loading, error, subscribeToMore } = useTenantQuery({
    variables: { tenantId: userId ?? '' },
    onCompleted: setData,
    fetchPolicy: 'network-only',
  });

  useEffect(() => {
    if (loading) return;
    subscribeToMore({
      variables: { tenantId: userId ?? '' },
      document: updatedTenant,
      updateQuery: (prev, { subscriptionData: { data } }) => {
        if (!data) return prev;
        setData(data);
        return data;
      },
    });
  }, [subscribeToMore, loading, userId]);

  return (
    <TenantContext.Provider value={{ data, loading, error }}>
      {children}
    </TenantContext.Provider>
  );
};

export default TenantContextProvider;
