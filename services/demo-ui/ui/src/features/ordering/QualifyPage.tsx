import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useCheckQualification, useAddCartItem } from './api';
import { AddressForm, type AddressFormData } from './AddressForm';
import { OfferingCard } from './OfferingCard';
import { PageLoader } from '../../design-system/components/common/PageLoader';

export default function QualifyPage() {
    const navigate = useNavigate();
    const { mutate: checkQualify, isPending, data: result, error, reset } = useCheckQualification();
    const { mutate: addToCart, isPending: isAddingToCart } = useAddCartItem();

    const sessionId = result?.sessionId ?? null;

    const [address, setAddress] = useState<AddressFormData>({
        street: '',
        number: '',
        city: '',
        zip: '',
    });

    const handleQualify = (e: React.FormEvent) => {
        e.preventDefault();
        checkQualify({ address });
    };

    const handleRecheck = () => {
        reset();
    };

    const handleSelectOffering = (offeringId: string) => {
        if (!sessionId) return;
        addToCart(
            { cartId: 'default-cart', offeringId, quantity: 1, qualificationSessionId: sessionId },
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
                <AddressForm
                    address={address}
                    onChange={setAddress}
                    onSubmit={handleQualify}
                    isPending={isPending}
                />
                {error && (
                    <div className="mt-4 space-y-2">
                        <p className="text-red-600">Failed to check qualification. Please try again.</p>
                        <button
                            onClick={handleRecheck}
                            className="bg-blue-600 text-white px-4 py-2 rounded shadow hover:bg-blue-700"
                        >
                            Retry
                        </button>
                    </div>
                )}
            </div>

            {result && result.status === 'Unqualified' && (
                <div className="bg-yellow-50 p-6 rounded-lg shadow border border-yellow-200 space-y-3">
                    <h2 className="text-xl font-semibold text-yellow-800">Service Not Available</h2>
                    <p className="text-yellow-700">
                        {result.unavailabilityReason ?? 'The service is not available at the provided address.'}
                    </p>
                    <button
                        onClick={handleRecheck}
                        className="bg-blue-600 text-white px-4 py-2 rounded shadow hover:bg-blue-700"
                    >
                        Check Another Address
                    </button>
                </div>
            )}

            {result && result.status === 'Qualified' && (
                <div className="space-y-4">
                    <h2 className="text-2xl font-bold">Available Offerings</h2>
                    {result.qualifiedOffers.length === 0 ? (
                        <p className="text-gray-500">No offerings available for this address.</p>
                    ) : (
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            {result.qualifiedOffers.map((off) => (
                                <OfferingCard
                                    key={off.offeringId}
                                    offering={off}
                                    onAddToCart={handleSelectOffering}
                                    isAddingToCart={isAddingToCart}
                                />
                            ))}
                        </div>
                    )}
                </div>
            )}

            {(isPending || isAddingToCart) && <PageLoader />}
        </div>
    );
}

