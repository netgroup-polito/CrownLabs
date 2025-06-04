import { createContext } from 'react';

interface IAuthContext {
  isLoggedIn: boolean;
  token?: string;
  userId?: string;
  logout: () => Promise<void>;
}

export const AuthContext = createContext<IAuthContext>({
  isLoggedIn: false,
  token: undefined,
  userId: undefined,
  logout: async () => void 0,
});
