import { getMainDefinition } from '@apollo/client/utilities';
import { GraphQLWsLink } from '@apollo/client/link/subscriptions';
import { createClient } from 'graphql-ws';
import {
  ApolloClient,
  ApolloProvider,
  HttpLink,
  InMemoryCache,
  split,
  type NormalizedCacheObject,
} from '@apollo/client';
import {
  type FC,
  type PropsWithChildren,
  useContext,
  useEffect,
  useState,
} from 'react';
import { VITE_APP_CROWNLABS_GRAPHQL_URL } from '../../env';
import { hasRenderingError } from '../../errorHandling/utils';
import { ErrorContext } from '../../errorHandling/ErrorContext';
import { useAuth } from 'react-oidc-context';

const httpUri = VITE_APP_CROWNLABS_GRAPHQL_URL;
const wsUri = httpUri.replace(/^http?/, 'ws') + '/subscription';
export interface Definition {
  kind: string;
  operation?: string;
}

const ApolloClientSetup: FC<PropsWithChildren> = props => {
  const { children } = props;
  const auth = useAuth();
  const token = auth.user?.id_token;
  const { errorsQueue } = useContext(ErrorContext);
  const [apolloClient, setApolloClient] =
    useState<ApolloClient<NormalizedCacheObject> | null>(null);

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
        }),
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
            httpLink,
          ),
          cache: new InMemoryCache(),
        }),
      );
    }
  }, [token]);

  return (
    <>
      <div>{auth.activeNavigator}</div>
      {(auth.isAuthenticated || hasRenderingError(errorsQueue)) &&
        apolloClient && (
          <ApolloProvider client={apolloClient}>{children}</ApolloProvider>
        )}
    </>
  );
};

export default ApolloClientSetup;
