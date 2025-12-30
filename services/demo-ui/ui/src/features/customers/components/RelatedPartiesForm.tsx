import { Plus, Trash2 } from 'lucide-react';
import type { RelatedParty } from '../types';

interface RelatedPartiesFormProps {
    items: Omit<RelatedParty, 'id'>[];
    onChange: (items: Omit<RelatedParty, 'id'>[]) => void;
}

export default function RelatedPartiesForm({ items, onChange }: RelatedPartiesFormProps) {
    const add = () => {
        onChange([...items, { relatedPartyId: '', role: '', name: '' }]);
    };

    const remove = (index: number) => {
        onChange(items.filter((_, i) => i !== index));
    };

    const update = (index: number, field: keyof Omit<RelatedParty, 'id'>, value: string) => {
        const updated = [...items];
        updated[index] = { ...updated[index], [field]: value };
        onChange(updated);
    };

    return (
        <div className="card form-section">
            <div className="section-header">
                <h3>Related Parties</h3>
                <button type="button" className="btn btn-secondary btn-sm" onClick={add}>
                    <Plus size={16} />
                    <span>Add Party</span>
                </button>
            </div>

            {items.length === 0 ? (
                <p className="empty-text">No related parties added</p>
            ) : (
                <div className="repeatable-list">
                    {items.map((item, index) => (
                        <div key={index} className="repeatable-item">
                            <div className="form-grid">
                                <div className="form-group">
                                    <label htmlFor={`rp-name-${index}`}>Party Name</label>
                                    <input
                                        id={`rp-name-${index}`}
                                        type="text"
                                        value={item.name}
                                        onChange={(e) => update(index, 'name', e.target.value)}
                                        placeholder="e.g. Parent Company Inc."
                                        required
                                    />
                                </div>
                                <div className="form-group">
                                    <label htmlFor={`rp-role-${index}`}>Role</label>
                                    <input
                                        id={`rp-role-${index}`}
                                        type="text"
                                        value={item.role}
                                        onChange={(e) => update(index, 'role', e.target.value)}
                                        placeholder="e.g. Parent, Partner"
                                        required
                                    />
                                </div>
                                <div className="form-group">
                                    <label htmlFor={`rp-id-${index}`}>Party ID</label>
                                    <input
                                        id={`rp-id-${index}`}
                                        type="text"
                                        value={item.relatedPartyId}
                                        onChange={(e) => update(index, 'relatedPartyId', e.target.value)}
                                        placeholder="Target Party UUID"
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
