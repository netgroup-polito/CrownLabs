import type { UserProfile } from 'oidc-client-ts';
import { createContext } from 'react';

interface IAuthContext {
  isLoggedIn: boolean;
  token?: string;
  userId?: string;
  profile?: UserProfile;
  logout: () => Promise<void>;
}

export const AuthContext = createContext<IAuthContext>({
  isLoggedIn: false,
  token: undefined,
  userId: undefined,
  profile: undefined,
  logout: async () => void 0,
});
