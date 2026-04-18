import { useState, useEffect, useRef, useCallback } from 'react';
import type { DebugMessage } from './types';
import { useAuth } from '../../auth/context';
import { getRuntimeConfig } from '../../config/runtime';

// Use correct WebSocket URL based on current origin
const getWebSocketUrl = () => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const cfg = getRuntimeConfig();
    const base = cfg.apiBaseUrl || cfg.apiUrl || window.location.origin;
    const url = new URL(base, window.location.origin);
    return `${protocol}//${url.host}/ws/debug`;
};

export function useDebugWebSocket() {
    const { getAccessTokenSilently, isAuthenticated } = useAuth();
    const [messages, setMessages] = useState<DebugMessage[]>([]);
    const [isConnected, setIsConnected] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const wsRef = useRef<WebSocket | null>(null);
    const connectRef = useRef<() => void>(undefined);

    const connect = useCallback(async () => {
        // Wait for authentication
        if (!isAuthenticated) {
            console.log('Not authenticated, skipping WebSocket connection');
            return;
        }

        try {
            // Get the access token from Auth0
            const token = await getAccessTokenSilently();
            const url = getWebSocketUrl();
            console.log('Connecting to Debug WebSocket:', url);

            // Pass JWT via Sec-WebSocket-Protocol header
            // Format: "access_token.BASE64_JWT"
            const ws = new WebSocket(url, [`access_token.${token}`]);

            ws.onopen = () => {
                console.log('Debug WebSocket connected');
                setIsConnected(true);
                setError(null);
            };

            ws.onmessage = (event) => {
                try {
                    const message: DebugMessage = JSON.parse(event.data);
                    // Add unique ID if missing
                    if (!message.id) {
                        message.id = crypto.randomUUID();
                    }
                    // Add timestamp if missing
                    if (!message.timestamp) {
                        message.timestamp = new Date().toISOString();
                    }

                    setMessages((prev) => {
                        const newMessages = [message, ...prev].slice(0, 100); // Keep last 100
                        return newMessages;
                    });
                } catch (err) {
                    console.error('Failed to parse websocket message:', err);
                }
            };

            ws.onclose = () => {
                setIsConnected(false);
                console.log('Debug WebSocket disconnected');
                // Auto-reconnect after 3s
                setTimeout(() => {
                    if (connectRef.current) {
                        connectRef.current();
                    }
                }, 3000);
            };

            ws.onerror = (e) => {
                console.error('WebSocket error:', e);
                setError('Connection error');
                ws.close();
            };

            wsRef.current = ws;
        } catch (err) {
            console.error('Failed to create WebSocket connection:', err);
            setError('Failed to create WebSocket connection');
            // Retry on connection error
            setTimeout(() => {
                if (connectRef.current) {
                    connectRef.current();
                }
            }, 3000);
        }
    }, [getAccessTokenSilently, isAuthenticated]);

    // Update ref whenever connect changes
    useEffect(() => {
        connectRef.current = connect;
    }, [connect]);

    useEffect(() => {
        // Defer connection to avoid synchronous state updates during effect execution
        const timer = setTimeout(() => {
            connect();
        }, 0);
        return () => {
            clearTimeout(timer);
            wsRef.current?.close();
        };
    }, [connect]);

    const clearMessages = () => setMessages([]);

    return {
        messages,
        isConnected,
        error,
        clearMessages,
    };
}
