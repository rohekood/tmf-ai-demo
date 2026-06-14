import { useState, useEffect, useRef, useCallback } from 'react';
import type { DebugMessage } from './types';
import { useAuth } from '../../auth/context';
import { getRuntimeConfig } from '../../config/runtime';

export type DebugConnectionStatus = 'connecting' | 'live' | 'reconnecting' | 'offline';

// Stop hammering the server forever; after this many consecutive failures we
// give up and let the user retry manually (prevents an infinite reconnect loop
// that floods the console with errors).
export const MAX_RECONNECT_ATTEMPTS = 6;
const BASE_RECONNECT_DELAY_MS = 1000;
const MAX_RECONNECT_DELAY_MS = 30000;

// Capped exponential backoff: 1s, 2s, 4s, 8s, 16s, 30s (capped).
export function getReconnectDelay(attempt: number): number {
    const delay = BASE_RECONNECT_DELAY_MS * 2 ** Math.max(0, attempt);
    return Math.min(MAX_RECONNECT_DELAY_MS, delay);
}

// Build the WebSocket URL from the configured API origin, falling back to the
// page origin. Always matches the page's ws/wss scheme.
export function getWebSocketUrl(): string {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const cfg = getRuntimeConfig();
    const base = cfg.apiBaseUrl || cfg.apiUrl || window.location.origin;
    const url = new URL(base, window.location.origin);
    return `${protocol}//${url.host}/ws/debug`;
}

export function useDebugWebSocket() {
    const { getAccessTokenSilently, isAuthenticated } = useAuth();
    const [messages, setMessages] = useState<DebugMessage[]>([]);
    const [status, setStatus] = useState<DebugConnectionStatus>('connecting');
    const [reconnectAttempt, setReconnectAttempt] = useState(0);
    const [error, setError] = useState<string | null>(null);

    const wsRef = useRef<WebSocket | null>(null);
    const connectRef = useRef<() => void>(undefined);
    const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
    const attemptRef = useRef(0);
    const errorLoggedRef = useRef(false);

    const scheduleReconnect = useCallback(() => {
        if (reconnectTimerRef.current) return; // a reconnect is already pending

        const attempt = attemptRef.current;
        if (attempt >= MAX_RECONNECT_ATTEMPTS) {
            setStatus('offline');
            return; // give up — user can retry manually
        }

        const delay = getReconnectDelay(attempt);
        attemptRef.current = attempt + 1;
        setReconnectAttempt(attemptRef.current);
        setStatus('reconnecting');

        reconnectTimerRef.current = setTimeout(() => {
            reconnectTimerRef.current = null;
            connectRef.current?.();
        }, delay);
    }, []);

    const connect = useCallback(async () => {
        if (!isAuthenticated) {
            return;
        }

        try {
            const token = await getAccessTokenSilently();
            const url = getWebSocketUrl();

            // Pass JWT via Sec-WebSocket-Protocol: "access_token.BASE64_JWT"
            const ws = new WebSocket(url, [`access_token.${token}`]);

            ws.onopen = () => {
                attemptRef.current = 0;
                errorLoggedRef.current = false;
                setReconnectAttempt(0);
                setStatus('live');
                setError(null);
            };

            ws.onmessage = (event) => {
                try {
                    const message: DebugMessage = JSON.parse(event.data);
                    if (!message.id) message.id = crypto.randomUUID();
                    if (!message.timestamp) message.timestamp = new Date().toISOString();
                    setMessages((prev) => [message, ...prev].slice(0, 100)); // keep last 100
                } catch (err) {
                    console.error('Failed to parse websocket message:', err);
                }
            };

            ws.onclose = () => {
                scheduleReconnect();
            };

            ws.onerror = () => {
                // Log only the first error of a failure streak to avoid console spam.
                if (!errorLoggedRef.current) {
                    console.error(`Debug WebSocket connection failed (${url}). Will retry with backoff.`);
                    errorLoggedRef.current = true;
                }
                setError('Connection error');
                ws.close(); // triggers onclose -> scheduleReconnect
            };

            wsRef.current = ws;
        } catch {
            if (!errorLoggedRef.current) {
                console.error('Failed to create Debug WebSocket connection. Will retry with backoff.');
                errorLoggedRef.current = true;
            }
            setError('Failed to create WebSocket connection');
            scheduleReconnect();
        }
    }, [getAccessTokenSilently, isAuthenticated, scheduleReconnect]);

    useEffect(() => {
        connectRef.current = connect;
    }, [connect]);

    useEffect(() => {
        const timer = setTimeout(() => connect(), 0);
        return () => {
            clearTimeout(timer);
            if (reconnectTimerRef.current) {
                clearTimeout(reconnectTimerRef.current);
                reconnectTimerRef.current = null;
            }
            wsRef.current?.close();
        };
    }, [connect]);

    const clearMessages = () => setMessages([]);

    // Manual retry after we've given up (status === 'offline').
    const reconnect = useCallback(() => {
        if (reconnectTimerRef.current) {
            clearTimeout(reconnectTimerRef.current);
            reconnectTimerRef.current = null;
        }
        attemptRef.current = 0;
        errorLoggedRef.current = false;
        setReconnectAttempt(0);
        setError(null);
        setStatus('connecting');
        connectRef.current?.();
    }, []);

    return {
        messages,
        status,
        reconnectAttempt,
        isConnected: status === 'live',
        error,
        clearMessages,
        reconnect,
    };
}
