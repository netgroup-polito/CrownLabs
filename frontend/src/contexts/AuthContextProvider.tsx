import { type FC, type PropsWithChildren, useContext, useEffect } from 'react';

import { ErrorContext } from '../errorHandling/ErrorContext';
import { ErrorTypes } from '../errorHandling/utils';
import { useAuth } from 'react-oidc-context';
import { AuthContext } from './AuthContext';

const AuthContextProvider: FC<PropsWithChildren> = props => {
  const { children } = props;
  const {
    isAuthenticated,
    isLoading,
    user,
    signinRedirect,
    removeUser,
    signoutRedirect,
  } = useAuth();
  const userId = user?.profile.preferred_username || '';
  const { makeErrorCatcher, setExecLogin, execLogin } =
    useContext(ErrorContext);

  useEffect(() => {
    if (execLogin && !isLoading && !isAuthenticated) {
      signinRedirect()
        .catch(makeErrorCatcher(ErrorTypes.KeycloakError))
        .finally(() => setExecLogin(false));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isAuthenticated, isLoading, setExecLogin, execLogin]);

  return (
    <AuthContext.Provider
      value={{
        isLoggedIn: isAuthenticated,
        token: user?.access_token,
        userId,
        logout: () => removeUser().then(() => signoutRedirect()),
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export default AuthContextProvider;
