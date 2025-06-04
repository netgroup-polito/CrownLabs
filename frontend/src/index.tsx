import React from 'react';
import ReactDOM from 'react-dom/client';
import './theming';
import '@ant-design/v5-patch-for-react-19';
import App from './App';
import TenantContextProvider from './contexts/TenantContextProvider';
import ErrorContextProvider from './errorHandling/ErrorContextProvider';
import AuthContextProvider from './contexts/AuthContextProvider';
import { AuthProvider, type AuthProviderProps } from 'react-oidc-context';
import {
  VITE_APP_CROWNLABS_OIDC_AUTHORITY,
  VITE_APP_CROWNLABS_OIDC_CLIENT_ID,
} from './env';
import { WebStorageStateStore } from 'oidc-client-ts';
import FullPageLoader from './components/common/FullPageLoader';

const oidcConfig: AuthProviderProps = {
  authority: VITE_APP_CROWNLABS_OIDC_AUTHORITY,
  client_id: VITE_APP_CROWNLABS_OIDC_CLIENT_ID,
  loadUserInfo: true,
  redirect_uri: window.location.origin,
  post_logout_redirect_uri: 'https://crownlabs.polito.it/',
  automaticSilentRenew: true,
  scope: 'openid profile email api',
  userStore: new WebStorageStateStore({ store: window.localStorage }),
};

const ApolloClientSetup = React.lazy(
  () => import('./graphql-components/apolloClientSetup/ApolloClientSetup'),
);

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ErrorContextProvider>
      <React.Suspense fallback={<FullPageLoader />}>
        <AuthProvider {...oidcConfig}>
          <AuthContextProvider>
            <ApolloClientSetup>
              <TenantContextProvider>
                <App />
              </TenantContextProvider>
            </ApolloClientSetup>
          </AuthContextProvider>
        </AuthProvider>
      </React.Suspense>
    </ErrorContextProvider>
  </React.StrictMode>,
);
