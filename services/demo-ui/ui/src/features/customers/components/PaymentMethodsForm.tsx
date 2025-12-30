import { Plus, Trash2 } from 'lucide-react';
import type { PaymentMethod } from '../types';

interface PaymentMethodsFormProps {
    items: Omit<PaymentMethod, 'id'>[];
    onChange: (items: Omit<PaymentMethod, 'id'>[]) => void;
}

export default function PaymentMethodsForm({ items, onChange }: PaymentMethodsFormProps) {
    const add = () => {
        onChange([...items, { type: '', token: '', isDefault: false, details: '{}' }]);
    };

    const remove = (index: number) => {
        onChange(items.filter((_, i) => i !== index));
    };

    const update = (index: number, field: keyof Omit<PaymentMethod, 'id'>, value: string | boolean) => {
        const updated = [...items];
        updated[index] = { ...updated[index], [field]: value };
        onChange(updated);
    };

    return (
        <div className="card form-section">
            <div className="section-header">
                <h3>Payment Methods</h3>
                <button type="button" className="btn btn-secondary btn-sm" onClick={add}>
                    <Plus size={16} />
                    <span>Add Method</span>
                </button>
            </div>

            {items.length === 0 ? (
                <p className="empty-text">No payment methods added</p>
            ) : (
                <div className="repeatable-list">
                    {items.map((item, index) => (
                        <div key={index} className="repeatable-item">
                            <div className="form-grid">
                                <div className="form-group">
                                    <label htmlFor={`pm-type-${index}`}>Type</label>
                                    <input
                                        id={`pm-type-${index}`}
                                        type="text"
                                        value={item.type}
                                        onChange={(e) => update(index, 'type', e.target.value)}
                                        placeholder="e.g. CreditCard"
                                        required
                                    />
                                </div>
                                <div className="form-group">
                                    <label htmlFor={`pm-token-${index}`}>Token</label>
                                    <input
                                        id={`pm-token-${index}`}
                                        type="text"
                                        value={item.token}
                                        onChange={(e) => update(index, 'token', e.target.value)}
                                        placeholder="Token"
                                        required
                                    />
                                </div>
                                <div className="form-group form-check-inline">
                                    <input
                                        id={`pm-default-${index}`}
                                        type="checkbox"
                                        checked={item.isDefault}
                                        onChange={(e) => update(index, 'isDefault', e.target.checked)}
                                    />
                                    <label htmlFor={`pm-default-${index}`}>Default</label>
                                </div>
                                <div className="form-group form-group--action">
                                    <button
                                        type="button"
                                        className="btn-icon btn-icon--danger"
                                        onClick={() => remove(index)}
                                        aria-label="Remove"
                                    >
                                        <Trash2 size={16} />
                                    </button>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}
