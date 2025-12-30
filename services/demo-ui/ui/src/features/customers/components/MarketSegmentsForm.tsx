import { Plus, Trash2 } from 'lucide-react';
import type { MarketSegment } from '../types';

interface MarketSegmentsFormProps {
    items: Omit<MarketSegment, 'id'>[];
    onChange: (items: Omit<MarketSegment, 'id'>[]) => void;
}

export default function MarketSegmentsForm({ items, onChange }: MarketSegmentsFormProps) {
    const add = () => {
        onChange([...items, { name: '', category: '' }]);
    };

    const remove = (index: number) => {
        onChange(items.filter((_, i) => i !== index));
    };

    const update = (index: number, field: keyof Omit<MarketSegment, 'id'>, value: string) => {
        const updated = [...items];
        updated[index] = { ...updated[index], [field]: value };
        onChange(updated);
    };

    return (
        <div className="card form-section">
            <div className="section-header">
                <h3>Market Segments</h3>
                <button type="button" className="btn btn-secondary btn-sm" onClick={add}>
                    <Plus size={16} />
                    <span>Add Segment</span>
                </button>
            </div>

            {items.length === 0 ? (
                <p className="empty-text">No market segments added</p>
            ) : (
                <div className="repeatable-list">
                    {items.map((item, index) => (
                        <div key={index} className="repeatable-item">
                            <div className="form-grid">
                                <div className="form-group">
                                    <label htmlFor={`ms-name-${index}`}>Segment Name</label>
                                    <input
                                        id={`ms-name-${index}`}
                                        type="text"
                                        value={item.name}
                                        onChange={(e) => update(index, 'name', e.target.value)}
                                        placeholder="e.g. Enterprise, SME"
                                        required
                                    />
                                </div>
                                <div className="form-group">
                                    <label htmlFor={`ms-cat-${index}`}>Category</label>
                                    <input
                                        id={`ms-cat-${index}`}
                                        type="text"
                                        value={item.category}
                                        onChange={(e) => update(index, 'category', e.target.value)}
                                        placeholder="e.g. B2B, B2C"
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
