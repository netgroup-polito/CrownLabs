import {
  type FC,
  type PropsWithChildren,
  useCallback,
  useContext,
  useEffect,
} from 'react';

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
        .catch(makeErrorCatcher(ErrorTypes.AuthError))
        .finally(() => setExecLogin(false));
    }
    if (isAuthenticated && execLogin) {
      setExecLogin(false);
    }
  }, [
    isAuthenticated,
    isLoading,
    setExecLogin,
    execLogin,
    signinRedirect,
    makeErrorCatcher,
  ]);

  const logout = useCallback(() => {
    return removeUser()
      .then(() => signoutRedirect())
      .catch(makeErrorCatcher(ErrorTypes.AuthError));
  }, [removeUser, signoutRedirect, makeErrorCatcher]);

  if (isLoading) return null;

  return (
    <AuthContext.Provider
      value={{
        isLoggedIn: isAuthenticated,
        token: user?.access_token,
        userId,
        profile: user?.profile,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export default AuthContextProvider;
