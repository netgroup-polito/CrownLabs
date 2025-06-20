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
import { AuthContext } from '../../contexts/AuthContext';
import FullPageLoader from '../../components/common/FullPageLoader';

import { loadErrorMessages, loadDevMessages } from '@apollo/client/dev';

if (import.meta.env.DEV) {
  loadDevMessages();
  loadErrorMessages();
}

const httpUri = VITE_APP_CROWNLABS_GRAPHQL_URL;
const wsUri = httpUri.replace(/^http?/, 'ws') + '/subscription';
export interface Definition {
  kind: string;
  operation?: string;
}

const ApolloClientSetup: FC<PropsWithChildren> = props => {
  const { children } = props;

  const { token, isLoggedIn } = useContext(AuthContext);
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

      const newClient = new ApolloClient({
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
      });

      setApolloClient(newClient);

      return () => {
        wsLink.client.dispose();
        newClient.clearStore();
      };
    }
  }, [token]);

  return (
    <>
      {(isLoggedIn || hasRenderingError(errorsQueue)) && apolloClient ? (
        <ApolloProvider client={apolloClient}>{children}</ApolloProvider>
      ) : (
        <FullPageLoader layoutWrap={true} />
      )}
    </>
  );
};

export default ApolloClientSetup;
