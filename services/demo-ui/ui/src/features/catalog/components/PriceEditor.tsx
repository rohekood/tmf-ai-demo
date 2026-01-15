import { Plus, Trash2, DollarSign } from 'lucide-react';
import type { ProductOfferingPrice } from '../types';
import { IconButton } from '../../../design-system/components/common/IconButton';

interface PriceEditorProps {
    prices: ProductOfferingPrice[];
    onChange: (prices: ProductOfferingPrice[]) => void;
}

export default function PriceEditor({ prices, onChange }: PriceEditorProps) {
    const handleAdd = () => {
        onChange([
            ...prices,
            {
                priceType: 'one_time',
                price: { unit: 'EUR', value: 0 },
            }
        ]);
    };

    const handleRemove = (index: number) => {
        onChange(prices.filter((_, i) => i !== index));
    };

    const handleChange = (index: number, updated: ProductOfferingPrice) => {
        const newPrices = [...prices];
        newPrices[index] = updated;
        onChange(newPrices);
    };

    return (
        <div className="card form-section">
            <div className="section-header">
                <h3>Pricing</h3>
                <button type="button" className="btn btn-secondary btn-sm" onClick={handleAdd}>
                    <Plus size={16} />
                    <span>Add Price</span>
                </button>
            </div>

            {prices.length === 0 ? (
                <p className="empty-text">No pricing defined for this offering.</p>
            ) : (
                <div className="repeatable-list">
                    {prices.map((p, index) => (
                        <div key={index} className="repeatable-item">
                            <div className="form-grid" style={{ gridTemplateColumns: '150px 1fr 120px auto', gap: '1rem' }}>
                                <div className="form-group">
                                    <label>Type</label>
                                    <select
                                        value={p.priceType}
                                        onChange={(e) => handleChange(index, { ...p, priceType: e.target.value as 'recurring' | 'one_time' | 'usage' })}
                                    >
                                        <option value="one_time">One Time</option>
                                        <option value="recurring">Recurring</option>
                                        <option value="usage">Usage</option>
                                    </select>
                                </div>
                                <div className="form-group">
                                    <label>Amount</label>
                                    <div className="input-with-icon">
                                        <DollarSign size={16} className="input-icon" />
                                        <input
                                            type="number"
                                            value={p.price.value}
                                            onChange={(e) => handleChange(index, { ...p, price: { ...p.price, value: parseFloat(e.target.value) || 0 } })}
                                            placeholder="0.00"
                                        />
                                    </div>
                                </div>
                                <div className="form-group">
                                    <label>Currency</label>
                                    <input
                                        type="text"
                                        value={p.price.unit}
                                        onChange={(e) => handleChange(index, { ...p, price: { ...p.price, unit: e.target.value } })}
                                        placeholder="EUR"
                                    />
                                </div>
                                <div className="form-group form-group--action">
                                    <IconButton
                                        variant="danger"
                                        size="sm"
                                        onClick={() => handleRemove(index)}
                                        icon={<Trash2 size={16} />}
                                    />
                                </div>
                            </div>
                            {p.priceType === 'recurring' && (
                                <div className="form-group mt-2">
                                    <label>Unit of Measure (e.g., month, year)</label>
                                    <input
                                        type="text"
                                        value={p.unitOfMeasure || ''}
                                        onChange={(e) => handleChange(index, { ...p, unitOfMeasure: e.target.value })}
                                        placeholder="month"
                                    />
                                </div>
                            )}
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}
