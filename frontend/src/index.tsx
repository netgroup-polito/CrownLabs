import React from 'react';
import ReactDOM from 'react-dom';
import './theming';
import App from './App';
import AuthContextProvider from './contexts/AuthContext';
import ApolloClientSetup from './graphql-components/apolloClientSetup/ApolloClientSetup';

ReactDOM.render(
  <React.StrictMode>
    <AuthContextProvider>
      <ApolloClientSetup>
        <App />
      </ApolloClientSetup>
    </AuthContextProvider>
  </React.StrictMode>,
  document.getElementById('root')
);
// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
// reportWebVitals();
