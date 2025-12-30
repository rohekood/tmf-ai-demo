import type { AppliedBillingRate } from '../types';

interface AppliedBillingRatesListProps {
    items: AppliedBillingRate[];
}

export default function AppliedBillingRatesList({ items }: AppliedBillingRatesListProps) {
    if (!items || items.length === 0) {
        return <p className="empty-text">No billing rates</p>;
    }

    return (
        <ul className="detail-list">
            {items.map((item) => (
                <li key={item.id} className="detail-item-row">
                    <div className="detail-row-content">
                        <span className="row-title">{item.productRef}</span>
                        <span className="row-subtitle">
                            {item.rateType}: {item.value}
                        </span>
                    </div>
                </li>
            ))}
        </ul>
    );
}
