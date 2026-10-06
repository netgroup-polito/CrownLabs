import type { UserProfile } from 'oidc-client-ts';
import { createContext } from 'react';

interface IAuthContext {
  isLoggedIn: boolean;
  token?: string;
  userId?: string;
  profile?: UserProfile;
  groups: string[];
  isClusterAdmin: boolean;
  isImagePublisher: boolean;
  logout: () => Promise<void>;
}

export const AuthContext = createContext<IAuthContext>({
  isLoggedIn: false,
  token: undefined,
  userId: undefined,
  profile: undefined,
  groups: [],
  isClusterAdmin: false,
  isImagePublisher: false,
  logout: async () => void 0,
});
