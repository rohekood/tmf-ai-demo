type RuntimeConfig = {
  auth0Domain: string;
  auth0ClientId: string;
  auth0Audience: string;
  apiUrl: string;
  apiBaseUrl: string;
};

type WindowWithConfig = Window & {
  __APP_CONFIG__?: {
    AUTH0_DOMAIN?: string;
    AUTH0_CLIENT_ID?: string;
    AUTH0_AUDIENCE?: string;
    API_URL?: string;
    API_BASE_URL?: string;
  };
};

export const getRuntimeConfig = (): RuntimeConfig => {
  const w = (typeof window !== 'undefined' ? window : undefined) as WindowWithConfig | undefined;
  const cfg = w?.__APP_CONFIG__;

  return {
    auth0Domain: cfg?.AUTH0_DOMAIN || import.meta.env.VITE_AUTH0_DOMAIN || '',
    auth0ClientId: cfg?.AUTH0_CLIENT_ID || import.meta.env.VITE_AUTH0_CLIENT_ID || '',
    auth0Audience: cfg?.AUTH0_AUDIENCE || import.meta.env.VITE_AUTH0_AUDIENCE || '',
    apiUrl: cfg?.API_URL || import.meta.env.VITE_API_URL || '',
    apiBaseUrl: cfg?.API_BASE_URL || import.meta.env.VITE_API_BASE_URL || '',
  };
};