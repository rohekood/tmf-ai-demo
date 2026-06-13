import React from 'react';

export interface AddressFormData {
    street: string;
    number: string;
    city: string;
    zip: string;
}

interface AddressFormProps {
    address: AddressFormData;
    onChange: (address: AddressFormData) => void;
    onSubmit: (e: React.FormEvent) => void;
    isPending: boolean;
}

export function AddressForm({ address, onChange, onSubmit, isPending }: AddressFormProps) {
    return (
        <form onSubmit={onSubmit} className="form">
            <div className="form-grid">
                <div className="form-field">
                    <label htmlFor="address-street" className="label">Street</label>
                    <input
                        id="address-street"
                        type="text"
                        value={address.street}
                        onChange={(e) => onChange({ ...address, street: e.target.value })}
                        required
                    />
                </div>
                <div className="form-field">
                    <label htmlFor="address-number" className="label">Number</label>
                    <input
                        id="address-number"
                        type="text"
                        value={address.number}
                        onChange={(e) => onChange({ ...address, number: e.target.value })}
                        required
                    />
                </div>
                <div className="form-field">
                    <label htmlFor="address-city" className="label">City</label>
                    <input
                        id="address-city"
                        type="text"
                        value={address.city}
                        onChange={(e) => onChange({ ...address, city: e.target.value })}
                        required
                    />
                </div>
                <div className="form-field">
                    <label htmlFor="address-zip" className="label">ZIP</label>
                    <input
                        id="address-zip"
                        type="text"
                        value={address.zip}
                        onChange={(e) => onChange({ ...address, zip: e.target.value })}
                        required
                    />
                </div>
            </div>
            <button type="submit" disabled={isPending} className="btn btn-primary">
                {isPending ? 'Checking...' : 'Check Availability'}
            </button>
        </form>
    );
}
