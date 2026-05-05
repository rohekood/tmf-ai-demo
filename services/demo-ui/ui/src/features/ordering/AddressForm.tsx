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
        <form onSubmit={onSubmit} className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
                <div>
                    <label htmlFor="address-street" className="block text-sm font-medium text-gray-700">Street</label>
                    <input
                        id="address-street"
                        type="text"
                        value={address.street}
                        onChange={(e) => onChange({ ...address, street: e.target.value })}
                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 p-2 border"
                        required
                    />
                </div>
                <div>
                    <label htmlFor="address-number" className="block text-sm font-medium text-gray-700">Number</label>
                    <input
                        id="address-number"
                        type="text"
                        value={address.number}
                        onChange={(e) => onChange({ ...address, number: e.target.value })}
                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 p-2 border"
                        required
                    />
                </div>
                <div>
                    <label htmlFor="address-city" className="block text-sm font-medium text-gray-700">City</label>
                    <input
                        id="address-city"
                        type="text"
                        value={address.city}
                        onChange={(e) => onChange({ ...address, city: e.target.value })}
                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 p-2 border"
                        required
                    />
                </div>
                <div>
                    <label htmlFor="address-zip" className="block text-sm font-medium text-gray-700">ZIP</label>
                    <input
                        id="address-zip"
                        type="text"
                        value={address.zip}
                        onChange={(e) => onChange({ ...address, zip: e.target.value })}
                        className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 p-2 border"
                        required
                    />
                </div>
            </div>
            <button
                type="submit"
                disabled={isPending}
                className="bg-blue-600 text-white px-4 py-2 rounded shadow hover:bg-blue-700 disabled:opacity-50"
            >
                {isPending ? 'Checking...' : 'Check Availability'}
            </button>
        </form>
    );
}
