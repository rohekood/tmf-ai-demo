import { describe, it, expect, afterEach, vi } from 'vitest';
import { getReconnectDelay, getWebSocketUrl, MAX_RECONNECT_ATTEMPTS } from './useDebugWebSocket';

describe('getReconnectDelay', () => {
    it('grows exponentially from 1s', () => {
        expect(getReconnectDelay(0)).toBe(1000);
        expect(getReconnectDelay(1)).toBe(2000);
        expect(getReconnectDelay(2)).toBe(4000);
        expect(getReconnectDelay(3)).toBe(8000);
        expect(getReconnectDelay(4)).toBe(16000);
    });

    it('caps at 30s', () => {
        expect(getReconnectDelay(5)).toBe(30000);
        expect(getReconnectDelay(10)).toBe(30000);
        expect(getReconnectDelay(MAX_RECONNECT_ATTEMPTS)).toBe(30000);
    });

    it('treats negative attempts as zero', () => {
        expect(getReconnectDelay(-3)).toBe(1000);
    });
});

describe('getWebSocketUrl', () => {
    afterEach(() => {
        vi.unstubAllEnvs();
    });

    it('derives ws host from the configured API origin', () => {
        vi.stubEnv('VITE_API_BASE_URL', 'http://api.example.com:8080');
        expect(getWebSocketUrl()).toBe('ws://api.example.com:8080/ws/debug');
    });

    it('falls back to the page origin when no API config is set', () => {
        // jsdom default origin is http://localhost:3000
        vi.stubEnv('VITE_API_BASE_URL', '');
        vi.stubEnv('VITE_API_URL', '');
        const url = getWebSocketUrl();
        expect(url).toMatch(/^ws:\/\/localhost(:\d+)?\/ws\/debug$/);
    });
});
