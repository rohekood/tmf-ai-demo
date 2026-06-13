import { useState, useEffect } from 'react';
import { User, Building2, Pencil } from 'lucide-react';
import type { PartyUnion } from '../types';
import PartySelector from '../PartySelector';
import { getPartyDisplayName } from '../types';
import './PartyPicker.css';

interface PartyPickerProps {
    value: PartyUnion | { id: string; name?: string; '@type'?: string } | null;
    onChange: (party: PartyUnion | null) => void;
    readOnly?: boolean;
    placeholder?: string;
    customActions?: React.ReactNode;
}

export default function PartyPicker({ value, onChange, readOnly = false, placeholder = 'Select a party...', customActions }: PartyPickerProps) {
    const [isSelecting, setIsSelecting] = useState(false);


    // If value is null, ensure we are in selection mode (unless readOnly which just shows empty)
    useEffect(() => {
        if (!value && !readOnly) {
            // eslint-disable-next-line react-hooks/set-state-in-effect
            setIsSelecting(true);
        }
    }, [value, readOnly]);

    const handleSelect = (party: PartyUnion) => {
        onChange(party);
        setIsSelecting(false);
    };

    // Helper to get display info safely
    const displayInfo = value ? {
        id: value.id,
        name: 'name' in value ? value.name : getPartyDisplayName(value as PartyUnion),
        type: '@type' in value ? value['@type'] : undefined
    } : null;

    if (isSelecting && !readOnly) {
        return (
            <div className="party-picker-selector">
                <PartySelector
                    selectedPartyId={value?.id}
                    onSelect={handleSelect}
                />
                <div className="party-picker-selector__actions">
                    <button
                        type="button"
                        className="btn btn-secondary btn--sm"
                        onClick={() => setIsSelecting(false)}
                    >
                        Cancel
                    </button>
                </div>
            </div>
        );
    }

    if (!value) {
        if (readOnly) return <div className="muted">No party selected</div>;

        return (
            <div className="party-picker-empty">
                <button
                    type="button"
                    className="btn btn-secondary"
                    onClick={() => setIsSelecting(true)}
                >
                    <User size={18} />
                    <span>{placeholder}</span>
                </button>
            </div>
        );
    }

    return (
        <div className="party-picker-card">
            <div className="party-picker-card__grid">
                <div className={`party-icon-circle ${displayInfo?.type?.toLowerCase() || 'unknown'}`}>
                    {displayInfo?.type === 'Individual' ? <User size={20} /> : <Building2 size={20} />}
                </div>

                <div className="party-picker-card__info">
                    <span className="party-picker-card__name">{displayInfo?.name || 'Unknown Party'}</span>

                    <div className="party-picker-card__meta">
                        <span style={{ fontWeight: 600 }}>ID:</span>
                        <span className="party-picker-card__meta-id" title={displayInfo?.id}>
                            {displayInfo?.id}
                        </span>
                    </div>

                    {displayInfo?.type && (
                        <div className="party-picker-card__meta">
                            <span style={{ fontWeight: 600 }}>Type:</span>
                            <span className={`badge badge-primary ${displayInfo.type.toLowerCase()}`}>
                                {displayInfo.type}
                            </span>
                        </div>
                    )}
                </div>

                {!readOnly && (
                    <div className="party-picker-card__actions">
                        {customActions}
                        <button
                            type="button"
                            className="btn btn-secondary btn--sm"
                            title="Change Party"
                            aria-label="Change Party"
                            onClick={() => setIsSelecting(true)}
                        >
                            <Pencil size={14} />
                            <span>Change</span>
                        </button>
                    </div>
                )}
            </div>
        </div>
    );
}
