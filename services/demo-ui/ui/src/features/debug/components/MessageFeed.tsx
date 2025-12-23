import type { DebugMessage } from '../types';
import { ArrowRight, Activity, Database, Terminal } from 'lucide-react';
import './MessageFeed.css';

interface MessageFeedProps {
    messages: DebugMessage[];
    selectedId: string | null;
    onSelect: (message: DebugMessage) => void;
}

export function MessageFeed({ messages, selectedId, onSelect }: MessageFeedProps) {
    if (messages.length === 0) {
        return (
            <div className="message-feed empty">
                <p>No messages captured yet.</p>
                <span>Perform actions in the app to see events here.</span>
            </div>
        );
    }

    return (
        <div className="message-feed" role="list">
            {messages.map((msg) => (
                <div
                    key={msg.id}
                    className={`feed-item ${msg.type} ${selectedId === msg.id ? 'selected' : ''}`}
                    onClick={() => onSelect(msg)}
                    role="listitem"
                >
                    <div className="feed-item-icon">
                        {msg.type === 'command' && <Terminal size={16} />}
                        {msg.type === 'event' && <Activity size={16} />}
                        {msg.type === 'query' && <Database size={16} />}
                        {msg.type === 'unknown' && <ArrowRight size={16} />}
                    </div>
                    <div className="feed-item-content">
                        <div className="feed-item-header">
                            <span className="feed-topic">{msg.topic}</span>
                            <span className="feed-time">
                                {new Date(msg.timestamp).toLocaleTimeString([], { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit', fractionalSecondDigits: 3 })}
                            </span>
                        </div>
                        <div className="feed-item-meta">
                            <span className="feed-service">{msg.service}</span>
                            {msg.correlationId && <span className="feed-correl">CID: {msg.correlationId.slice(-8)}</span>}
                        </div>
                    </div>
                </div>
            ))}
        </div>
    );
}
