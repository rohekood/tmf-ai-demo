import { createContext, useContext } from 'react';

type AuthUser = {
  name?: string;
  email?: string;
  picture?: string;
};

type LoginOptions = {
  authorizationParams?: {
    screen_hint?: string;
  };
};

type LogoutOptions = {
  logoutParams?: {
    returnTo?: string;
  };
};

export type AuthDisabledReason = 'insecure-origin' | 'not-configured' | null;

export type AuthClient = {
  enabled: boolean;
  /** When `enabled` is false, why auth is unavailable. */
  disabledReason?: AuthDisabledReason;
  isAuthenticated: boolean;
  isLoading: boolean;
  user?: AuthUser;
  loginWithRedirect: (options?: LoginOptions) => Promise<void>;
  logout: (options?: LogoutOptions) => void;
  getAccessTokenSilently: () => Promise<string>;
};

const insecureOriginMessage = 'Auth0 is disabled on insecure origins. Use HTTPS or localhost.';

export const fallbackAuthClient: AuthClient = {
  enabled: false,
  isAuthenticated: false,
  isLoading: false,
  user: undefined,
  loginWithRedirect: async () => {
    console.warn(insecureOriginMessage);
  },
  logout: () => {
    console.warn(insecureOriginMessage);
  },
  getAccessTokenSilently: async () => {
    throw new Error(insecureOriginMessage);
  },
};

export const AuthContext = createContext<AuthClient>(fallbackAuthClient);

export function useAuth() {
  return useContext(AuthContext);
}
