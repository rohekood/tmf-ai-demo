import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { isAxiosError } from 'axios';
import { useCheckQualification, useAddCartItem } from './api';
import { AddressForm, type AddressFormData } from './AddressForm';
import { OfferingCard } from './OfferingCard';
import { PageLoader } from '../../design-system/components/common/PageLoader';
import { useNotification } from '../../design-system/components/common/Toast';
import { useAuth } from '../../auth/context';
import { CART_ID_KEY, QUALIFY_RESUME_KEY } from './storage';
import './ordering.css';

export default function QualifyPage() {
    const navigate = useNavigate();
    const { showToast } = useNotification();
    const { isAuthenticated, loginWithRedirect } = useAuth();
    const { mutate: checkQualify, isPending, data: result, error, reset } = useCheckQualification();
    const { mutate: addToCart, isPending: isAddingToCart } = useAddCartItem();
    const [sessionExpired, setSessionExpired] = useState(false);

    const sessionId = result?.sessionId ?? null;

    // Restore the address entered before login (if any) as the initial form
    // value, so a returning user sees it prefilled.
    const [address, setAddress] = useState<AddressFormData>(() => {
        const saved = sessionStorage.getItem(QUALIFY_RESUME_KEY);
        if (saved) {
            try {
                return JSON.parse(saved) as AddressFormData;
            } catch {
                // Ignore malformed resume state.
            }
        }
        return { street: '', number: '', city: '', zip: '' };
    });

    // After returning from login, re-run qualification for the restored address
    // so the user resumes with customer pricing. The add-to-cart itself is not
    // auto-retried — the customer must confirm against the re-priced offers.
    useEffect(() => {
        if (!isAuthenticated) return;
        const saved = sessionStorage.getItem(QUALIFY_RESUME_KEY);
        if (!saved) return;
        sessionStorage.removeItem(QUALIFY_RESUME_KEY);
        try {
            checkQualify({ address: JSON.parse(saved) as AddressFormData });
        } catch {
            // Ignore malformed resume state.
        }
    }, [isAuthenticated, checkQualify]);

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

        // Adding to cart requires an authenticated user. Anonymous visitors are
        // sent into the login flow; we stash the address so qualification can be
        // resumed (and re-priced for the customer) when they come back.
        if (!isAuthenticated) {
            sessionStorage.setItem(QUALIFY_RESUME_KEY, JSON.stringify(address));
            void loginWithRedirect({
                appState: { returnTo: '/order/qualify' },
                authorizationParams: { screen_hint: 'login' },
            });
            return;
        }

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
        <div className="page page--narrow">
            <div className="page-header">
                <div>
                    <h1 className="page-title">Service Qualification</h1>
                    <p className="page-subtitle">Enter your address to see available services.</p>
                </div>
            </div>

            {sessionExpired && (
                <div role="alert" className="alert alert-warning">
                    <span>
                        Session expired – please{' '}
                        <button onClick={handleRecheck} className="btn-link">re-check availability</button>.
                    </span>
                    <button onClick={() => setSessionExpired(false)} aria-label="Dismiss" className="btn-link">✕</button>
                </div>
            )}

            <div className="card">
                <AddressForm
                    address={address}
                    onChange={setAddress}
                    onSubmit={handleQualify}
                    isPending={isPending}
                />
                {error && (
                    <div className="stack-sm" style={{ marginTop: '1rem' }}>
                        <p className="field-error">Failed to check qualification. Please try again.</p>
                        <button onClick={handleRecheck} className="btn btn-primary" style={{ alignSelf: 'flex-start' }}>
                            Retry
                        </button>
                    </div>
                )}
            </div>

            {result && result.status === 'Unqualified' && (
                <div className="card stack">
                    <h2 className="section-title">Service Not Available</h2>
                    <p className="muted">
                        {result.unavailabilityReason ?? 'The service is not available at the provided address.'}
                    </p>
                    <button onClick={handleRecheck} className="btn btn-primary" style={{ alignSelf: 'flex-start' }}>
                        Check Another Address
                    </button>
                </div>
            )}

            {result && result.status === 'Qualified' && (
                <div className="stack">
                    <h2 className="section-title">Available Offerings</h2>
                    {result.qualifiedOffers.length === 0 ? (
                        <p className="muted">No offerings available for this address.</p>
                    ) : (
                        <div className="grid-auto">
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
