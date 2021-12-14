import Keycloak from 'keycloak-js';
import {
  createContext,
  FC,
  PropsWithChildren,
  useContext,
  useEffect,
  useState,
} from 'react';
import {
  REACT_APP_CROWNLABS_OIDC_PROVIDER_URL,
  REACT_APP_CROWNLABS_OIDC_CLIENT_ID,
  REACT_APP_CROWNLABS_OIDC_REALM,
} from '../env';
import { ErrorContext } from '../errorHandling/ErrorContext';
import { ErrorTypes } from '../errorHandling/utils';
interface IAuthContext {
  isLoggedIn: boolean;
  token?: string;
  userId?: string;
}

export const AuthContext = createContext<IAuthContext>({
  isLoggedIn: false,
  token: undefined,
  userId: undefined,
});

const kc = Keycloak({
  url: REACT_APP_CROWNLABS_OIDC_PROVIDER_URL,
  realm: REACT_APP_CROWNLABS_OIDC_REALM,
  clientId: REACT_APP_CROWNLABS_OIDC_CLIENT_ID,
});

export const logout = () => kc.logout({ redirectUri: window.location.origin });

const AuthContextProvider: FC<PropsWithChildren<{}>> = props => {
  const { children } = props;
  const [isLoggedIn, setIsLoggedIn] = useState<boolean>(false);
  const [userId, setUserId] = useState<undefined | string>(undefined);
  const [token, setToken] = useState<undefined | string>(undefined);
  const { makeErrorCatcher, setExecLogin, execLogin } =
    useContext(ErrorContext);

  useEffect(() => {
    if (execLogin) {
      kc.init({ onLoad: 'login-required' })
        .then((authenticated: boolean) => {
          if (authenticated) {
            setIsLoggedIn(true);
            setToken(kc.idToken);
          } else {
            setIsLoggedIn(false);
            setToken(undefined);
            setUserId(undefined);
          }
          kc.loadUserInfo()
            .then((res: any) => setUserId(res.preferred_username))
            .catch(makeErrorCatcher(ErrorTypes.KeycloakError));
        })
        .catch(makeErrorCatcher(ErrorTypes.KeycloakError))
        .finally(() => setExecLogin(false));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [setExecLogin, execLogin]);

  return (
    <AuthContext.Provider
      value={{
        isLoggedIn,
        token,
        userId,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export default AuthContextProvider;
