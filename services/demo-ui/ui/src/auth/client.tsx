import { useMemo, type ReactNode } from 'react';
import { Auth0Provider, useAuth0 } from '@auth0/auth0-react';
import { AuthContext, fallbackAuthClient, type AuthClient, type AuthDisabledReason } from './context';
import { getRuntimeConfig } from '../config/runtime';

const isSecureOrigin = () => {
  if (typeof window === 'undefined') {
    return true;
  }
  const host = window.location.hostname;
  return window.isSecureContext || host === 'localhost' || host === '127.0.0.1';
};

const isAuthConfigured = () => {
  const cfg = getRuntimeConfig();
  return Boolean(cfg.auth0Domain && cfg.auth0ClientId && cfg.auth0Audience);
};

const isAuthEnabled = () => isSecureOrigin() && isAuthConfigured();

const getDisabledReason = (): AuthDisabledReason => {
  if (!isSecureOrigin()) return 'insecure-origin';
  if (!isAuthConfigured()) return 'not-configured';
  return null;
};

function AuthBridge({ children }: { children: ReactNode }) {
  const auth0 = useAuth0();

  const authClient = useMemo<AuthClient>(() => ({
    enabled: true,
    disabledReason: null,
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
  const cfg = getRuntimeConfig();
  const domain = cfg.auth0Domain;
  const clientId = cfg.auth0ClientId;
  const audience = cfg.auth0Audience;

  if (!isAuthEnabled()) {
    if (!isSecureOrigin()) {
      console.warn(`Auth0 disabled for origin ${window.location.origin}. Use HTTPS or localhost.`);
    } else if (!isAuthConfigured()) {
      console.warn('Auth0 configuration missing, check AUTH0_* runtime env vars.');
    }

    return (
      <AuthContext.Provider value={{ ...fallbackAuthClient, disabledReason: getDisabledReason() }}>
        {children}
      </AuthContext.Provider>
    );
  }

  return (
    <Auth0Provider
      domain={domain}
      clientId={clientId}
      authorizationParams={{
        redirect_uri: window.location.origin,
        audience,
      }}
      // Persist the session across full page reloads. Without this the default
      // in-memory cache drops the session on refresh and silent re-auth fails,
      // bouncing the user back to the login screen.
      cacheLocation="localstorage"
      useRefreshTokens={true}
    >
      <AuthBridge>{children}</AuthBridge>
    </Auth0Provider>
  );
}
