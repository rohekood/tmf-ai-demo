import type { CustomerInteraction } from '../types';
import { formatDateTime } from '../../../lib/date';

interface InteractionsListProps {
    items: CustomerInteraction[];
}

export default function InteractionsList({ items }: InteractionsListProps) {
    if (!items || items.length === 0) {
        return <p className="empty-text">No interactions logged</p>;
    }

    return (
        <div className="interactions-timeline">
            {items.map((item) => (
                <div key={item.id} className="timeline-item">
                    <div className="timeline-header">
                        <span className="timeline-date">{formatDateTime(item.interactionDate)}</span>
                        <span className="timeline-agent">by {item.agentId}</span>
                    </div>
                    <div className="timeline-content">
                        <strong>{item.type} ({item.channel})</strong>
                        <p>{item.description}</p>
                    </div>
                </div>
            ))}
            <style>{`
                .interactions-timeline {
                    display: flex;
                    flex-direction: column;
                    gap: 1rem;
                }
                .timeline-item {
                    border-left: 2px solid var(--border-color);
                    padding-left: 1rem;
                }
                .timeline-header {
                    font-size: 0.85rem;
                    color: var(--text-muted);
                    margin-bottom: 0.25rem;
                }
                .timeline-date {
                    margin-right: 0.5rem;
                }
                .timeline-content p {
                    margin: 0.25rem 0 0;
                }
            `}</style>
        </div>
    );
}
