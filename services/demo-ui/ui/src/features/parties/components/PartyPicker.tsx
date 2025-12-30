import { useState, useEffect } from 'react';
import { User, Building2, Pencil } from 'lucide-react';
import type { PartyUnion } from '../types';
import PartySelector from '../PartySelector';
import { getPartyDisplayName } from '../types';

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
                <div className="mt-2 text-right">
                    <button
                        type="button"
                        className="btn btn-sm btn-secondary"
                        onClick={() => setIsSelecting(false)}
                    >
                        Cancel
                    </button>
                </div>
            </div>
        );
    }

    if (!value) {
        if (readOnly) return <div className="text-muted">No party selected</div>;

        return (
            <div className="party-picker-empty">
                <button
                    type="button"
                    className="btn btn-secondary w-100 d-flex align-items-center justify-content-center gap-2 p-3"
                    onClick={() => setIsSelecting(true)}
                >
                    <User size={18} />
                    <span>{placeholder}</span>
                </button>
            </div>
        );
    }

    return (
        <div className="party-picker-card border rounded p-3 bg-light mb-0">
            <div style={{ display: 'grid', gridTemplateColumns: 'auto 1fr auto', gap: '1rem', alignItems: 'center' }}>
                <div className={`party-icon-circle ${displayInfo?.type?.toLowerCase() || 'unknown'}`}>
                    {displayInfo?.type === 'Individual' ? <User size={20} /> : <Building2 size={20} />}
                </div>

                <div className="d-flex flex-row align-items-center flex-wrap gap-3 overflow-hidden">
                    <span className="fw-bold fs-5 text-nowrap">{displayInfo?.name || 'Unknown Party'}</span>

                    <div className="d-flex align-items-center text-muted small text-nowrap">
                        <span className="fw-semibold me-1">ID:</span>
                        <span className="font-monospace text-truncate" style={{ maxWidth: '150px' }} title={displayInfo?.id}>
                            {displayInfo?.id}
                        </span>
                    </div>

                    {displayInfo?.type && (
                        <div className="d-flex align-items-center text-nowrap">
                            <span className="text-muted small fw-semibold me-1">Type:</span>
                            <span className={`badge badge-outline ${displayInfo.type.toLowerCase()}`}>
                                {displayInfo.type}
                            </span>
                        </div>
                    )}
                </div>

                {!readOnly && (
                    <div className="d-flex flex-row align-items-center gap-2 text-nowrap">
                        {customActions}
                        <button
                            type="button"
                            className="btn btn-sm btn-outline-secondary d-flex align-items-center"
                            title="Change Party"
                            aria-label="Change Party"
                            onClick={() => setIsSelecting(true)}
                        >
                            <Pencil size={14} className="me-2" />
                            <span>Change</span>
                        </button>
                    </div>
                )}
            </div>
        </div>
    );
}
