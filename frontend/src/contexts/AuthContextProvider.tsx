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
import {
  VITE_APP_CROWNLABS_GROUPS_ADMIN_CLAIM,
  VITE_APP_CROWNLABS_GROUPS_CLAIM_PREFIX,
} from '../env';

const AuthContextProvider: FC<PropsWithChildren> = props => {
  const { children } = props;
  const {
    isAuthenticated,
    isLoading,
    user,
    signinRedirect,
    removeUser,
    signoutRedirect,
    startSilentRenew,
  } = useAuth();

  const { makeErrorCatcher, setExecLogin, execLogin } =
    useContext(ErrorContext);

  const loginErrorCatcher = makeErrorCatcher(ErrorTypes.AuthError);

  useEffect(() => {
    if (isAuthenticated) {
      startSilentRenew();
    }
  }, [startSilentRenew, isAuthenticated]);

  useEffect(() => {
    if (!isLoading && (!isAuthenticated || execLogin)) {
      signinRedirect().catch(loginErrorCatcher);
      setExecLogin(false);
    }
  }, [
    execLogin,
    setExecLogin,
    isLoading,
    isAuthenticated,
    signinRedirect,
    loginErrorCatcher,
  ]);

  const logout = useCallback(() => {
    return removeUser()
      .then(() => signoutRedirect())
      .catch(loginErrorCatcher);
  }, [removeUser, signoutRedirect, loginErrorCatcher]);

  const groups = (user?.profile.groups ?? []) as string[];
  const isClusterAdmin = groups.includes(
    `${VITE_APP_CROWNLABS_GROUPS_CLAIM_PREFIX}:${VITE_APP_CROWNLABS_GROUPS_ADMIN_CLAIM}`,
  );
  const isImagePublisher = groups.includes(
    `${VITE_APP_CROWNLABS_GROUPS_CLAIM_PREFIX}:image-publisher`,
  );

  return isLoading ? null : (
    <AuthContext.Provider
      value={{
        isLoggedIn: isAuthenticated,
        token: user?.id_token,
        userId: user?.profile.preferred_username || '',
        profile: user?.profile,
        groups,
        isClusterAdmin,
        isImagePublisher,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export default AuthContextProvider;
