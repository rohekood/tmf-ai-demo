import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useCheckQualification, useAddCartItem } from './api';
import { PageLoader } from '../../design-system/components/common/PageLoader';

export default function QualifyPage() {
    const navigate = useNavigate();
    const { mutate: checkQualify, isPending, data: session, error } = useCheckQualification();
    const { mutate: addToCart, isPending: isAddingToCart } = useAddCartItem();

    const [address, setAddress] = useState({
        street: 'Pärnu mnt 1',
        city: 'Tallinn',
        postcode: '10148',
        country: 'EE'
    });

    const handleQualify = (e: React.FormEvent) => {
        e.preventDefault();
        checkQualify({ address });
    };

    const handleSelectOffering = (offeringId: string) => {
        if (!session) return;
        addToCart(
            { cartId: 'default-cart', offeringId, quantity: 1, qualificationSessionId: session.id },
            {
                onSuccess: () => navigate('/order/cart'),
            }
        );
    };

    return (
        <div className="p-8 max-w-4xl mx-auto space-y-8">
            <div>
                <h1 className="text-3xl font-bold">Service Qualification</h1>
                <p className="text-gray-500 mt-2">Enter your address to see available services.</p>
            </div>

            <div className="bg-white p-6 rounded-lg shadow border border-gray-100">
                <form onSubmit={handleQualify} className="space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700">Street</label>
                            <input
                                type="text"
                                value={address.street}
                                onChange={(e) => setAddress({ ...address, street: e.target.value })}
                                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 p-2 border"
                                required
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700">City</label>
                            <input
                                type="text"
                                value={address.city}
                                onChange={(e) => setAddress({ ...address, city: e.target.value })}
                                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 p-2 border"
                                required
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700">Postcode</label>
                            <input
                                type="text"
                                value={address.postcode}
                                onChange={(e) => setAddress({ ...address, postcode: e.target.value })}
                                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 p-2 border"
                                required
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700">Country</label>
                            <input
                                type="text"
                                value={address.country}
                                onChange={(e) => setAddress({ ...address, country: e.target.value })}
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
                {error && <div className="mt-4 text-red-600">Failed to check qualification.</div>}
            </div>

            {session && (
                <div className="space-y-4">
                    <h2 className="text-2xl font-bold">Available Offerings</h2>
                    {session.qualifiedOfferings.length === 0 ? (
                        <p className="text-gray-500">No offerings available for this address.</p>
                    ) : (
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            {session.qualifiedOfferings.map((off) => (
                                <div key={off.offeringId} className="bg-white p-6 rounded-lg shadow border border-gray-100 flex flex-col justify-between">
                                    <div>
                                        <h3 className="text-lg font-bold">{off.name}</h3>
                                        {off.description && <p className="text-gray-500 text-sm mt-1">{off.description}</p>}
                                        <p className="text-2xl font-semibold mt-4">
                                            {off.price} {off.currency}
                                        </p>
                                    </div>
                                    <button
                                        onClick={() => handleSelectOffering(off.offeringId)}
                                        disabled={isAddingToCart}
                                        className="mt-6 bg-green-600 text-white px-4 py-2 rounded shadow hover:bg-green-700 disabled:opacity-50 w-full"
                                    >
                                        Add to Cart
                                    </button>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            )}
            {(isPending || isAddingToCart) && <PageLoader />}
        </div>
    );
}
