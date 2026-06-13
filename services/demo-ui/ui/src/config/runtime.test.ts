import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { getRuntimeConfig } from './runtime';

type MutableWindow = Window & {
  __APP_CONFIG__?: Record<string, string | undefined>;
};

describe('getRuntimeConfig', () => {
  const w = window as MutableWindow;

  beforeEach(() => {
    delete w.__APP_CONFIG__;
    vi.stubEnv('VITE_AUTH0_DOMAIN', 'tenant.eu.auth0.com');
    vi.stubEnv('VITE_AUTH0_CLIENT_ID', 'client-123');
    vi.stubEnv('VITE_AUTH0_AUDIENCE', 'http://localhost/api');
    vi.stubEnv('VITE_API_URL', 'http://localhost');
  });

  afterEach(() => {
    delete w.__APP_CONFIG__;
    vi.unstubAllEnvs();
  });

  it('uses VITE_* env vars when __APP_CONFIG__ is absent', () => {
    const cfg = getRuntimeConfig();
    expect(cfg.auth0Domain).toBe('tenant.eu.auth0.com');
    expect(cfg.auth0ClientId).toBe('client-123');
    expect(cfg.auth0Audience).toBe('http://localhost/api');
  });

  it('falls back to VITE_* env vars when __APP_CONFIG__ values are empty strings', () => {
    // This mirrors local `yarn dev`, where public/config.js injects empty strings.
    w.__APP_CONFIG__ = {
      AUTH0_DOMAIN: '',
      AUTH0_CLIENT_ID: '',
      AUTH0_AUDIENCE: '',
      API_URL: '',
      API_BASE_URL: '',
    };

    const cfg = getRuntimeConfig();
    expect(cfg.auth0Domain).toBe('tenant.eu.auth0.com');
    expect(cfg.auth0ClientId).toBe('client-123');
    expect(cfg.auth0Audience).toBe('http://localhost/api');
  });

  it('prefers non-empty __APP_CONFIG__ values over VITE_* env vars', () => {
    w.__APP_CONFIG__ = {
      AUTH0_DOMAIN: 'runtime.eu.auth0.com',
      AUTH0_CLIENT_ID: 'runtime-client',
      AUTH0_AUDIENCE: 'https://prod/api',
    };

    const cfg = getRuntimeConfig();
    expect(cfg.auth0Domain).toBe('runtime.eu.auth0.com');
    expect(cfg.auth0ClientId).toBe('runtime-client');
    expect(cfg.auth0Audience).toBe('https://prod/api');
  });
});
