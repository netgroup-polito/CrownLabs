import { FC, PropsWithChildren, useEffect, useState } from 'react';
import { useContext } from 'react';
import { AuthContext } from '../../contexts/AuthContext';
import { ApolloProvider } from '@apollo/client';

import {
  ApolloClient,
  createHttpLink,
  InMemoryCache,
  NormalizedCacheObject,
} from '@apollo/client';
import { setContext } from '@apollo/client/link/context';
import { REACT_APP_CROWNLABS_GRAPHQL_URL } from '../../env';

const ApolloClientSetup: FC<PropsWithChildren<{}>> = props => {
  const { children } = props;
  const { token, isLoggedIn } = useContext(AuthContext);
  const [apolloClient, setApolloClient] = useState<
    ApolloClient<NormalizedCacheObject> | undefined
  >(undefined);

  useEffect(() => {
    const httpLink = createHttpLink({
      uri: 'https://' + REACT_APP_CROWNLABS_GRAPHQL_URL,
    });

    const authLink = setContext((_, { headers }) => {
      return {
        headers: {
          ...headers,
          authorization: token ? `Bearer ${token}` : '',
        },
      };
    });

    setApolloClient(
      new ApolloClient({
        link: authLink.concat(httpLink),
        cache: new InMemoryCache(),
      })
    );
  }, [token]);
  return (
    <>
      {isLoggedIn && apolloClient && (
        <ApolloProvider client={apolloClient}>{children}</ApolloProvider>
      )}
    </>
  );
};

export default ApolloClientSetup;
