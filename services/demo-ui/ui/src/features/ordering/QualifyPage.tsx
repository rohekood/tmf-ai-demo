import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { isAxiosError } from 'axios';
import { useCheckQualification, useAddCartItem } from './api';
import { AddressForm, type AddressFormData } from './AddressForm';
import { OfferingCard } from './OfferingCard';
import { PageLoader } from '../../design-system/components/common/PageLoader';
import { useNotification } from '../../design-system/components/common/Toast';
import { CART_ID_KEY } from './storage';

export default function QualifyPage() {
    const navigate = useNavigate();
    const { showToast } = useNotification();
    const { mutate: checkQualify, isPending, data: result, error, reset } = useCheckQualification();
    const { mutate: addToCart, isPending: isAddingToCart } = useAddCartItem();
    const [sessionExpired, setSessionExpired] = useState(false);

    const sessionId = result?.sessionId ?? null;

    const [address, setAddress] = useState<AddressFormData>({
        street: '',
        number: '',
        city: '',
        zip: '',
    });

    const handleQualify = (e: React.FormEvent) => {
        e.preventDefault();
        setSessionExpired(false);
        checkQualify({ address });
    };

    const handleRecheck = () => {
        setSessionExpired(false);
        reset();
        checkQualify({ address });
    };

    const handleSelectOffering = (offeringId: string) => {
        if (!sessionId) return;
        const existingCartId = localStorage.getItem(CART_ID_KEY) || undefined;
        addToCart(
            { cartId: existingCartId, offeringId, quantity: 1, qualificationSessionId: sessionId },
            {
                onSuccess: (response) => {
                    if (response?.cartId) {
                        localStorage.setItem(CART_ID_KEY, response.cartId);
                    }
                    navigate('/order/cart');
                },
                onError: (err) => {
                    if (isAxiosError(err) && err.response?.status === 422) {
                        const errorCode = err.response.data?.error;
                        if (errorCode === 'SESSION_EXPIRED') {
                            setSessionExpired(true);
                        } else {
                            showToast('Not available at your address', 'error');
                        }
                    } else {
                        showToast('Failed to add item – try again', 'error');
                    }
                },
            }
        );
    };

    return (
        <div className="p-8 max-w-4xl mx-auto space-y-8">
            <div>
                <h1 className="text-3xl font-bold">Service Qualification</h1>
                <p className="text-gray-500 mt-2">Enter your address to see available services.</p>
            </div>

            {sessionExpired && (
                <div role="alert" className="bg-yellow-50 border border-yellow-300 text-yellow-800 px-4 py-3 rounded flex items-center justify-between">
                    <span>Session expired – please <button onClick={handleRecheck} className="underline font-medium">re-check availability</button>.</span>
                    <button onClick={() => setSessionExpired(false)} aria-label="Dismiss" className="ml-4 text-yellow-600 hover:text-yellow-800">✕</button>
                </div>
            )}

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
