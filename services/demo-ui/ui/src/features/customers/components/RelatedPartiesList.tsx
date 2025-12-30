import type { RelatedParty } from '../types';

interface RelatedPartiesListProps {
    items: RelatedParty[];
}

export default function RelatedPartiesList({ items }: RelatedPartiesListProps) {
    if (!items || items.length === 0) {
        return <p className="empty-text">No related parties</p>;
    }

    return (
        <ul className="detail-list">
            {items.map((item) => (
                <li key={item.id} className="detail-item-row">
                    <div className="detail-row-content">
                        <span className="row-title">{item.name}</span>
                        <span className="row-subtitle">{item.role}</span>
                    </div>
                </li>
            ))}
        </ul>
    );
}
