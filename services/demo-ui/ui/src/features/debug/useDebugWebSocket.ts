import { useState, useEffect, useRef, useCallback } from 'react';
import type { DebugMessage } from './types';

// Use correct WebSocket URL based on current origin
const getWebSocketUrl = () => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    // Use BFF URL if running on different port, otherwise current origin
    const bffUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
    const url = new URL(bffUrl);
    return `${protocol}//${url.host}/ws/debug`;
};

export function useDebugWebSocket() {
    const [messages, setMessages] = useState<DebugMessage[]>([]);
    const [isConnected, setIsConnected] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const wsRef = useRef<WebSocket | null>(null);

    const connectRef = useRef<() => void>(undefined);

    const connect = useCallback(() => {
        try {
            const url = getWebSocketUrl();
            console.log('Connecting to Debug WebSocket:', url);
            const ws = new WebSocket(url);

            ws.onopen = () => {
                console.log('Debug WebSocket connected');
                setIsConnected(true);
                setError(null); // Keep this to clear error on successful connection
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
                setError('Connection error'); // Keep this to show error
                ws.close();
            };

            wsRef.current = ws;
        } catch (err) {
            console.error('Failed to create WebSocket connection:', err);
            setError('Failed to create WebSocket connection'); // Keep this to show error
            // Retry on connection error
            setTimeout(() => {
                if (connectRef.current) {
                    connectRef.current();
                }
            }, 3000);
        }
    }, []);

    // Update ref whenever connect changes (which is stable here, but good practice)
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
