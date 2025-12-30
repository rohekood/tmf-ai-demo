import { Plus, Trash2 } from 'lucide-react';
import type { AppliedBillingRate } from '../types';

interface AppliedBillingRatesFormProps {
    items: Omit<AppliedBillingRate, 'id'>[];
    onChange: (items: Omit<AppliedBillingRate, 'id'>[]) => void;
}

export default function AppliedBillingRatesForm({ items, onChange }: AppliedBillingRatesFormProps) {
    const add = () => {
        onChange([...items, { productRef: '', rateType: '', value: 0 }]);
    };

    const remove = (index: number) => {
        onChange(items.filter((_, i) => i !== index));
    };

    const update = (index: number, field: keyof Omit<AppliedBillingRate, 'id'>, value: string | number) => {
        const updated = [...items];
        updated[index] = { ...updated[index], [field]: value };
        onChange(updated);
    };

    return (
        <div className="card form-section">
            <div className="section-header">
                <h3>Applied Billing Rates</h3>
                <button type="button" className="btn btn-secondary btn-sm" onClick={add}>
                    <Plus size={16} />
                    <span>Add Rate</span>
                </button>
            </div>

            {items.length === 0 ? (
                <p className="empty-text">No billing rates added</p>
            ) : (
                <div className="repeatable-list">
                    {items.map((item, index) => (
                        <div key={index} className="repeatable-item">
                            <div className="form-grid">
                                <div className="form-group">
                                    <label htmlFor={`br-ref-${index}`}>Product/Asset Ref</label>
                                    <input
                                        id={`br-ref-${index}`}
                                        type="text"
                                        value={item.productRef}
                                        onChange={(e) => update(index, 'productRef', e.target.value)}
                                        placeholder="Product or Asset ID"
                                        required
                                    />
                                </div>
                                <div className="form-group">
                                    <label htmlFor={`br-type-${index}`}>Rate Type</label>
                                    <input
                                        id={`br-type-${index}`}
                                        type="text"
                                        value={item.rateType}
                                        onChange={(e) => update(index, 'rateType', e.target.value)}
                                        placeholder="e.g. Discount, Override"
                                        required
                                    />
                                </div>
                                <div className="form-group">
                                    <label htmlFor={`br-val-${index}`}>Value</label>
                                    <input
                                        id={`br-val-${index}`}
                                        type="number"
                                        value={item.value}
                                        onChange={(e) => update(index, 'value', parseFloat(e.target.value) || 0)}
                                        required
                                    />
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
