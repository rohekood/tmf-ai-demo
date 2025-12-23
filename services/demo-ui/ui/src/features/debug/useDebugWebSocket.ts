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

    const connect = useCallback(() => {
        try {
            const url = getWebSocketUrl();
            const ws = new WebSocket(url);

            ws.onopen = () => {
                setIsConnected(true);
                setError(null);
                console.log('Debug WebSocket connected');
            };

            ws.onmessage = (event) => {
                try {
                    const msg = JSON.parse(event.data);
                    // Handle single message or array (if batched/buffered)
                    if (Array.isArray(msg)) {
                        setMessages((prev) => [...msg, ...prev].slice(0, 1000)); // Keep last 1000
                    } else {
                        setMessages((prev) => [msg, ...prev].slice(0, 1000));
                    }
                } catch (e) {
                    console.error('Failed to parse websocket message:', e);
                }
            };

            ws.onclose = () => {
                setIsConnected(false);
                console.log('Debug WebSocket disconnected');
                // Auto-reconnect after 3s
                setTimeout(connect, 3000);
            };

            ws.onerror = (e) => {
                console.error('WebSocket error:', e);
                setError('Connection error');
                ws.close();
            };

            wsRef.current = ws;
        } catch (e) {
            console.error('Failed to create WebSocket:', e);
            setError('Failed to create WebSocket connection');
        }
    }, []);

    useEffect(() => {
        connect();
        return () => {
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
