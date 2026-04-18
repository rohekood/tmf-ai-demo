import { useMemo, type ReactNode } from 'react';
import { Auth0Provider, useAuth0 } from '@auth0/auth0-react';
import { AuthContext, fallbackAuthClient, type AuthClient } from './context';

const isSecureOrigin = () => {
  if (typeof window === 'undefined') {
    return true;
  }
  return window.isSecureContext || window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';
};

const isAuthConfigured = () => {
  return Boolean(import.meta.env.VITE_AUTH0_DOMAIN && import.meta.env.VITE_AUTH0_CLIENT_ID && import.meta.env.VITE_AUTH0_AUDIENCE);
};

const isAuthEnabled = () => isSecureOrigin() && isAuthConfigured();

function AuthBridge({ children }: { children: ReactNode }) {
  const auth0 = useAuth0();

  const authClient = useMemo<AuthClient>(() => ({
    enabled: true,
    isAuthenticated: auth0.isAuthenticated,
    isLoading: auth0.isLoading,
    user: auth0.user,
    loginWithRedirect: auth0.loginWithRedirect,
    logout: auth0.logout,
    getAccessTokenSilently: auth0.getAccessTokenSilently,
  }), [auth0]);

  return <AuthContext.Provider value={authClient}>{children}</AuthContext.Provider>;
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const domain = import.meta.env.VITE_AUTH0_DOMAIN;
  const clientId = import.meta.env.VITE_AUTH0_CLIENT_ID;
  const audience = import.meta.env.VITE_AUTH0_AUDIENCE;

  if (!isAuthEnabled()) {
    if (!isSecureOrigin()) {
      console.warn(`Auth0 disabled for origin ${window.location.origin}. Use HTTPS or localhost.`);
    } else if (!isAuthConfigured()) {
      console.warn('Auth0 configuration missing, check VITE_AUTH0_* env vars.');
    }

    return <AuthContext.Provider value={fallbackAuthClient}>{children}</AuthContext.Provider>;
  }

  return (
    <Auth0Provider
      domain={domain}
      clientId={clientId}
      authorizationParams={{
        redirect_uri: window.location.origin,
        audience,
      }}
    >
      <AuthBridge>{children}</AuthBridge>
    </Auth0Provider>
  );
}
