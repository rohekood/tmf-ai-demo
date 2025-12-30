import type { PaymentMethod } from '../types';
import { CreditCard } from 'lucide-react';

interface PaymentMethodsListProps {
    items: PaymentMethod[];
}

export default function PaymentMethodsList({ items }: PaymentMethodsListProps) {
    if (!items || items.length === 0) {
        return <p className="empty-text">No payment methods</p>;
    }

    return (
        <ul className="detail-list">
            {items.map((item) => (
                <li key={item.id} className="detail-item-row">
                    <div className="icon-row">
                        <CreditCard size={18} className="text-muted" />
                        <div className="detail-row-content">
                            <span className="row-title">{item.type}</span>
                            <span className="row-subtitle">{item.token}</span>
                        </div>
                        {item.isDefault && <span className="status-badge active">Default</span>}
                    </div>
                </li>
            ))}
        </ul>
    );
}
