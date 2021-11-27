import { getMainDefinition } from '@apollo/client/utilities';
import { ApolloProvider } from '@apollo/react-hooks';
import { InMemoryCache } from 'apollo-cache-inmemory';
import { ApolloClient } from 'apollo-client';
import { ApolloLink, split } from 'apollo-link';
import { HttpLink } from 'apollo-link-http';
import { WebSocketLink } from 'apollo-link-ws';
import { FC, PropsWithChildren, useContext, useEffect, useState } from 'react';
import { AuthContext } from '../../contexts/AuthContext';
import { REACT_APP_CROWNLABS_GRAPHQL_URL } from '../../env';

const httpUri = REACT_APP_CROWNLABS_GRAPHQL_URL;
const wsUri = httpUri.replace(/^http?/, 'ws') + '/subscription';
export interface Definition {
  kind: string;
  operation?: string;
}

const ApolloClientSetup: FC<PropsWithChildren<{}>> = props => {
  const { children } = props;
  const { token, isLoggedIn } = useContext(AuthContext);
  const [apolloClient, setApolloClient] = useState<any>('');

  useEffect(() => {
    if (token) {
      const httpLink = new HttpLink({
        uri: httpUri,
        headers: {
          authorization: token ? `Bearer ${token}` : '',
        },
      });

      const wsLink = new WebSocketLink({
        uri: wsUri,
        options: {
          // Automatic reconnect in case of connection error
          reconnect: true,
          connectionParams: {
            authorization: token ? `Bearer ${token}` : '',
          },
        },
      });

      const terminatingLink = split(
        ({ query }) => {
          const { kind, operation }: Definition = getMainDefinition(query);
          // If this is a subscription query, use wsLink, otherwise use httpLink
          return kind === 'OperationDefinition' && operation === 'subscription';
        },
        wsLink,
        httpLink
      );

      const link = ApolloLink.from([terminatingLink]);

      setApolloClient(
        new ApolloClient({
          link,
          cache: new InMemoryCache(),
        })
      );
    }
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
