import { getMainDefinition } from '@apollo/client/utilities';
import { GraphQLWsLink } from '@apollo/client/link/subscriptions';
import { createClient } from 'graphql-ws';
import {
  ApolloClient,
  ApolloProvider,
  HttpLink,
  InMemoryCache,
  split,
} from '@apollo/client';
import { FC, PropsWithChildren, useContext, useEffect, useState } from 'react';
import { AuthContext } from '../../contexts/AuthContext';
import { REACT_APP_CROWNLABS_GRAPHQL_URL } from '../../env';
import { hasRenderingError } from '../../errorHandling/utils';
import { ErrorContext } from '../../errorHandling/ErrorContext';

const httpUri = REACT_APP_CROWNLABS_GRAPHQL_URL;
const wsUri = httpUri.replace(/^http?/, 'ws') + '/subscription';
export interface Definition {
  kind: string;
  operation?: string;
}

const ApolloClientSetup: FC<PropsWithChildren<{}>> = props => {
  const { children } = props;
  const { token, isLoggedIn } = useContext(AuthContext);
  const { errorsQueue } = useContext(ErrorContext);
  const [apolloClient, setApolloClient] = useState<any>('');

  useEffect(() => {
    if (token) {
      const authHeader = {
        authorization: `Bearer ${token}`,
      };
      const httpLink = new HttpLink({
        uri: httpUri,
        headers: authHeader,
      });

      const wsLink = new GraphQLWsLink(
        createClient({
          url: wsUri,
          connectionParams: authHeader,
          shouldRetry: () => true,
        })
      );

      setApolloClient(
        new ApolloClient({
          link: split(
            ({ query }) => {
              const { kind, operation }: Definition = getMainDefinition(query);
              // If this is a subscription query, use wsLink, otherwise use httpLink
              return (
                kind === 'OperationDefinition' && operation === 'subscription'
              );
            },
            wsLink,
            httpLink
          ),
          cache: new InMemoryCache(),
        })
      );
    }
  }, [token]);
  return (
    <>
      {(isLoggedIn || hasRenderingError(errorsQueue)) && apolloClient && (
        <ApolloProvider client={apolloClient}>{children}</ApolloProvider>
      )}
    </>
  );
};

export default ApolloClientSetup;
