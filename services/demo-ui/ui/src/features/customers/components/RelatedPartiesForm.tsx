import { Plus, Trash2 } from 'lucide-react';
import type { RelatedParty } from '../types';
import PartyPicker from '../../parties/components/PartyPicker';
import { getPartyDisplayName } from '../../parties/types';

interface RelatedPartiesFormProps {
    items: Partial<RelatedParty>[];
    onChange: (items: Partial<RelatedParty>[]) => void;
}

export default function RelatedPartiesForm({ items, onChange }: RelatedPartiesFormProps) {
    const add = () => {
        onChange([...items, { relatedPartyId: '', role: '', name: '' }]);
    };

    const remove = (index: number) => {
        onChange(items.filter((_, i) => i !== index));
    };

    const update = (index: number, field: keyof RelatedParty, value: string) => {
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
                                <div className="form-group form-group--full">
                                    <label>Party</label>
                                    <PartyPicker
                                        value={item.relatedPartyId ? {
                                            id: item.relatedPartyId,
                                            name: item.name
                                        } : null}
                                        onChange={(party) => {
                                            const updated = [...items];
                                            if (party) {
                                                updated[index] = {
                                                    ...updated[index],
                                                    relatedPartyId: party.id,
                                                    name: getPartyDisplayName(party)
                                                };
                                            } else {
                                                updated[index] = { ...updated[index], relatedPartyId: '', name: '' };
                                            }
                                            onChange(updated);
                                        }}
                                        placeholder="Select related party..."
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
