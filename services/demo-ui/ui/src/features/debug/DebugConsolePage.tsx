import { useState, useMemo } from 'react';
import { useDebugWebSocket } from './useDebugWebSocket';
import { DebugFilters } from './components/DebugFilters';
import { MessageFeed } from './components/MessageFeed';
import { MessageDetail } from './components/MessageDetail';
import type { DebugFilterState } from './types';
import { Wifi, WifiOff } from 'lucide-react';
import './DebugConsolePage.css';

export default function DebugConsolePage() {
    const { messages, isConnected, clearMessages } = useDebugWebSocket();
    const [selectedId, setSelectedId] = useState<string | null>(null);

    const [filter, setFilter] = useState<DebugFilterState>({
        search: '',
        services: [],
        types: [],
    });

    const filteredMessages = useMemo(() => {
        return messages.filter((msg) => {
            // Search filter
            if (filter.search) {
                const searchLower = filter.search.toLowerCase();
                const matchesTopic = msg.topic.toLowerCase().includes(searchLower);
                const matchesPayload = JSON.stringify(msg.payload).toLowerCase().includes(searchLower);
                if (!matchesTopic && !matchesPayload) return false;
            }

            // Service filter
            if (filter.services.length > 0) {
                const msgService = msg.service;
                // Map unknown services to what they likely are if possible, or just exact match
                // Our consumer sets service from topic prefix "tmf.{service}"
                if (!filter.services.some(s => msgService.includes(s))) return false;
            }

            // Type filter
            if (filter.types.length > 0) {
                if (!filter.types.includes(msg.type)) return false;
            }

            return true;
        });
    }, [messages, filter]);

    const selectedMessage = useMemo(
        () => messages.find((m) => m.id === selectedId) || null,
        [messages, selectedId]
    );

    return (
        <div className="debug-console">
            <div className="console-header">
                <div className="header-left">
                    <h2>Debug Console</h2>
                    <div className={`connection-status ${isConnected ? 'connected' : 'disconnected'}`} role="status">
                        {isConnected ? <Wifi size={16} /> : <WifiOff size={16} />}
                        <span>{isConnected ? 'Live' : 'Disconnected'}</span>
                    </div>
                </div>
                <div className="header-right">
                    <span className="message-count">{messages.length} events captured</span>
                </div>
            </div>

            <div className="console-layout">
                {/* Left Sidebar: Filters */}
                <div className="console-sidebar">
                    <DebugFilters
                        filter={filter}
                        onChange={setFilter}
                        onClear={clearMessages}
                        totalCount={messages.length}
                        filteredCount={filteredMessages.length}
                    />
                </div>

                {/* Middle: Feed */}
                <div className="console-feed">
                    <MessageFeed
                        messages={filteredMessages}
                        selectedId={selectedId}
                        onSelect={(msg) => setSelectedId(msg.id)}
                    />
                </div>

                {/* Right: Details */}
                <div className="console-details">
                    <MessageDetail
                        message={selectedMessage}
                        onClose={() => setSelectedId(null)}
                    />
                </div>
            </div>
        </div>
    );
}
