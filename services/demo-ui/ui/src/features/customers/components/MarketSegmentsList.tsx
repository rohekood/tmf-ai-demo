import type { MarketSegment } from '../types';

interface MarketSegmentsListProps {
    items: MarketSegment[];
}

export default function MarketSegmentsList({ items }: MarketSegmentsListProps) {
    if (!items || items.length === 0) {
        return <p className="empty-text">No market segments</p>;
    }

    return (
        <div className="tags-list">
            {items.map((item) => (
                <span key={item.id} className="tag-badge">
                    {item.name} ({item.category})
                </span>
            ))}
        </div>
    );
}
