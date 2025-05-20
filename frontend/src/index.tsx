import React from 'react';
import ReactDOM from 'react-dom/client';
import './theming';
import '@ant-design/v5-patch-for-react-19';
import App from './App';
import ApolloClientSetup from './graphql-components/apolloClientSetup/ApolloClientSetup';
import TenantContextProvider from './contexts/TenantContextProvider';
import ErrorContextProvider from './errorHandling/ErrorContextProvider';
import AuthContextProvider from './contexts/AuthContextProvider';
import { AuthProvider, type AuthProviderProps } from 'react-oidc-context';
import {
  VITE_APP_CROWNLABS_OIDC_AUTHORITY,
  VITE_APP_CROWNLABS_OIDC_CLIENT_ID,
} from './env';
import { WebStorageStateStore } from 'oidc-client-ts';

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

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ErrorContextProvider>
      <AuthProvider {...oidcConfig}>
        <AuthContextProvider>
          <ApolloClientSetup>
            <TenantContextProvider>
              <App />
            </TenantContextProvider>
          </ApolloClientSetup>
        </AuthContextProvider>
      </AuthProvider>
    </ErrorContextProvider>
  </React.StrictMode>,
);
// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
// reportWebVitals();
