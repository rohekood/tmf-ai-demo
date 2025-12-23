import type { DebugMessage } from '../types';
import { Copy, X, Terminal, ArrowRightLeft } from 'lucide-react';
import './MessageDetail.css';

interface MessageDetailProps {
    message: DebugMessage | null;
    onClose: () => void;
}

export function MessageDetail({ message, onClose }: MessageDetailProps) {
    if (!message) {
        return (
            <div className="message-detail empty">
                <Terminal size={48} />
                <p>Select a message to view details</p>
            </div>
        );
    }

    const handleCopy = () => {
        navigator.clipboard.writeText(JSON.stringify(message.payload, null, 2));
    };

    return (
        <div className="message-detail">
            <div className="detail-header">
                <div className="detail-title">
                    <span className={`type-badge ${message.type}`}>{message.type}</span>
                    <span className="detail-topic">{message.topic}</span>
                </div>
                <div className="detail-actions">
                    <button className="btn-icon" onClick={handleCopy} title="Copy Payload">
                        <Copy size={16} />
                    </button>
                    <button className="btn-icon" onClick={onClose} title="Close">
                        <X size={16} />
                    </button>
                </div>
            </div>

            <div className="detail-content">
                <div className="meta-grid">
                    <div className="meta-item">
                        <label>Timestamp</label>
                        <span className="mono">{new Date(message.timestamp).toISOString()}</span>
                    </div>
                    <div className="meta-item">
                        <label>Service</label>
                        <span>{message.service}</span>
                    </div>
                    <div className="meta-item">
                        <label>ID</label>
                        <span className="mono">{message.id}</span>
                    </div>
                    {message.correlationId && (
                        <div className="meta-item">
                            <label>Correlation ID</label>
                            <span className="mono">{message.correlationId}</span>
                        </div>
                    )}
                    {message.replyTo && (
                        <div className="meta-item">
                            <label>Reply To</label>
                            <span className="mono"><ArrowRightLeft size={12} /> {message.replyTo}</span>
                        </div>
                    )}
                </div>

                <div className="payload-section">
                    <div className="section-title">Payload</div>
                    <pre className="json-viewer">
                        {JSON.stringify(message.payload, null, 2)}
                    </pre>
                </div>
            </div>
        </div>
    );
}
