import { Plus, Trash2 } from 'lucide-react';
import type { ProductSpecCharacteristic } from '../types';

interface CharacteristicEditorProps {
    characteristics: Record<string, ProductSpecCharacteristic>;
    onChange: (characteristics: Record<string, ProductSpecCharacteristic>) => void;
}

export default function CharacteristicEditor({ characteristics, onChange }: CharacteristicEditorProps) {
    const handleAdd = () => {
        const name = `New Characteristic ${Object.keys(characteristics).length + 1}`;
        onChange({
            ...characteristics,
            [name]: {
                name,
                valueType: 'string',
                configurable: true,
            },
        });
    };

    const handleRemove = (name: string) => {
        const updated = { ...characteristics };
        delete updated[name];
        onChange(updated);
    };

    const handleChange = (oldName: string, updated: ProductSpecCharacteristic) => {
        const newCharacteristics = { ...characteristics };
        if (oldName !== updated.name) {
            delete newCharacteristics[oldName];
        }
        newCharacteristics[updated.name] = updated;
        onChange(newCharacteristics);
    };

    return (
        <div className="card form-section">
            <div className="section-header">
                <h3>Characteristics</h3>
                <button type="button" className="btn btn-secondary btn-sm" onClick={handleAdd}>
                    <Plus size={16} />
                    <span>Add Characteristic</span>
                </button>
            </div>

            {Object.keys(characteristics).length === 0 ? (
                <p className="empty-text">No characteristics defined.</p>
            ) : (
                <div className="repeatable-list">
                    {Object.entries(characteristics).map(([name, char]) => (
                        <div key={name} className="repeatable-item">
                            <div className="form-grid" style={{ gridTemplateColumns: '1fr 1fr 120px auto', gap: '1rem' }}>
                                <div className="form-group">
                                    <label>Name</label>
                                    <input
                                        type="text"
                                        value={char.name}
                                        onChange={(e) => handleChange(name, { ...char, name: e.target.value })}
                                        required
                                    />
                                </div>
                                <div className="form-group">
                                    <label>Value Type</label>
                                    <select
                                        value={char.valueType}
                                        onChange={(e) => handleChange(name, { ...char, valueType: e.target.value as 'string' | 'number' | 'boolean' })}
                                    >
                                        <option value="string">String</option>
                                        <option value="number">Number</option>
                                        <option value="boolean">Boolean</option>
                                    </select>
                                </div>
                                <div className="form-group" style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                                    <label className="checkbox-label">
                                        <input
                                            type="checkbox"
                                            checked={char.configurable}
                                            onChange={(e) => handleChange(name, { ...char, configurable: e.target.checked })}
                                        />
                                        <span>Configurable</span>
                                    </label>
                                </div>
                                <div className="form-group form-group--action">
                                    <button
                                        type="button"
                                        className="btn-icon btn-icon--danger"
                                        onClick={() => handleRemove(name)}
                                    >
                                        <Trash2 size={16} />
                                    </button>
                                </div>
                            </div>
                            <div className="form-group mt-2">
                                <label>Description (Optional)</label>
                                <input
                                    type="text"
                                    value={char.description || ''}
                                    placeholder="Briefly describe this characteristic..."
                                    onChange={(e) => handleChange(name, { ...char, description: e.target.value })}
                                />
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}
