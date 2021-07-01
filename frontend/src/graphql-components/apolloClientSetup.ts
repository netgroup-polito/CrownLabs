import { ApolloClient, createHttpLink, InMemoryCache } from '@apollo/client';
import { setContext } from '@apollo/client/link/context';
import { REACT_APP_CROWNLABS_GRAPHQL_URL } from '../env';

const TOKEN = '';

const httpLink = createHttpLink({
  uri: REACT_APP_CROWNLABS_GRAPHQL_URL,
});

const authLink = setContext((_, { headers }) => {
  const token = TOKEN;
  return {
    headers: {
      ...headers,
      authorization: token ? `Bearer ${token}` : '',
    },
  };
});

const client = new ApolloClient({
  link: authLink.concat(httpLink),
  cache: new InMemoryCache(),
});

export { client };
