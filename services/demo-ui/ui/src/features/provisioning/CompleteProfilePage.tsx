import React, { useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useQueryClient } from '@tanstack/react-query';
import { useAuth } from '../../auth/context';
import { useNotification } from '../../design-system/components/common/Toast';
import { useProvisionCustomer, meCustomerQueryKey, type ProvisionRequest } from './api';

// Best-effort split of a single display name into given/family parts.
function splitName(name?: string): { givenName: string; familyName: string } {
    if (!name) return { givenName: '', familyName: '' };
    const parts = name.trim().split(/\s+/);
    if (parts.length === 1) return { givenName: parts[0], familyName: '' };
    return { givenName: parts[0], familyName: parts.slice(1).join(' ') };
}

export default function CompleteProfilePage() {
    const navigate = useNavigate();
    const location = useLocation();
    const queryClient = useQueryClient();
    const { showToast } = useNotification();
    const { user } = useAuth();
    const { mutate: provision, isPending } = useProvisionCustomer();

    const initialName = splitName(user?.name);
    const [form, setForm] = useState<ProvisionRequest>({
        givenName: initialName.givenName,
        familyName: initialName.familyName,
        phone: '',
        street: '',
        city: '',
        postcode: '',
        country: '',
    });

    const returnTo = (location.state as { returnTo?: string } | null)?.returnTo ?? '/';

    const update = (field: keyof ProvisionRequest, value: string) =>
        setForm((f) => ({ ...f, [field]: value }));

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        // Email is NOT sent: the backend derives identity from the verified
        // access-token claim. The email field below is display-only.
        provision(form, {
            onSuccess: (result) => {
                // Seed the resolve cache directly with the provisioned result so
                // the gate passes immediately — without depending on a re-resolve
                // by email (which requires the Auth0 email claim, see A1).
                queryClient.setQueryData(meCustomerQueryKey, result);
                showToast('Profile saved', 'success');
                navigate(returnTo, { replace: true });
            },
            onError: () => showToast('Failed to save profile – try again', 'error'),
        });
    };

    return (
        <div className="page page--narrow">
            <div className="page-header">
                <div>
                    <h1 className="page-title">Complete your profile</h1>
                    <p className="page-subtitle">
                        We need a few details to set up your customer account before you continue.
                    </p>
                </div>
            </div>

            <div className="card">
                <form onSubmit={handleSubmit} className="form">
                    <div className="form-field">
                        <label htmlFor="profile-email" className="label">Email</label>
                        <input id="profile-email" type="email" value={user?.email ?? ''} readOnly disabled />
                    </div>
                    <div className="form-grid">
                        <div className="form-field">
                            <label htmlFor="profile-given" className="label">First name</label>
                            <input
                                id="profile-given"
                                type="text"
                                value={form.givenName}
                                onChange={(e) => update('givenName', e.target.value)}
                                required
                            />
                        </div>
                        <div className="form-field">
                            <label htmlFor="profile-family" className="label">Last name</label>
                            <input
                                id="profile-family"
                                type="text"
                                value={form.familyName}
                                onChange={(e) => update('familyName', e.target.value)}
                                required
                            />
                        </div>
                        <div className="form-field">
                            <label htmlFor="profile-phone" className="label">Phone</label>
                            <input
                                id="profile-phone"
                                type="tel"
                                value={form.phone}
                                onChange={(e) => update('phone', e.target.value)}
                            />
                        </div>
                        <div className="form-field">
                            <label htmlFor="profile-street" className="label">Street</label>
                            <input
                                id="profile-street"
                                type="text"
                                value={form.street}
                                onChange={(e) => update('street', e.target.value)}
                            />
                        </div>
                        <div className="form-field">
                            <label htmlFor="profile-city" className="label">City</label>
                            <input
                                id="profile-city"
                                type="text"
                                value={form.city}
                                onChange={(e) => update('city', e.target.value)}
                            />
                        </div>
                        <div className="form-field">
                            <label htmlFor="profile-postcode" className="label">Postcode</label>
                            <input
                                id="profile-postcode"
                                type="text"
                                value={form.postcode}
                                onChange={(e) => update('postcode', e.target.value)}
                            />
                        </div>
                        <div className="form-field">
                            <label htmlFor="profile-country" className="label">Country</label>
                            <input
                                id="profile-country"
                                type="text"
                                value={form.country}
                                onChange={(e) => update('country', e.target.value)}
                            />
                        </div>
                    </div>
                    <button type="submit" disabled={isPending} className="btn btn-primary">
                        {isPending ? 'Saving...' : 'Save and continue'}
                    </button>
                </form>
            </div>
        </div>
    );
}
