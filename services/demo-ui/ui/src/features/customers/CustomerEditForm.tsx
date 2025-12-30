import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Save, Plus, Trash2, Loader2 } from 'lucide-react';
import { useUpdateCustomer } from './api';
import type { Customer, CustomerStatus, PrivacyConsent, CreditProfile, CustomerAccount, RelatedParty } from './types';
import PartyPicker from '../parties/components/PartyPicker';
import RelatedPartiesForm from './components/RelatedPartiesForm';
import { getPartyDisplayName } from '../parties/types';
import '../parties/PartyFormPage.css';

interface CustomerEditFormProps {
    customer: Customer;
}

export default function CustomerEditForm({ customer }: CustomerEditFormProps) {
    const navigate = useNavigate();
    const updateMutation = useUpdateCustomer();

    const [status, setStatus] = useState<CustomerStatus>(customer.status);
    const [name, setName] = useState(customer.name);

    const [privacyConsents, setPrivacyConsents] = useState<Partial<PrivacyConsent>[]>(customer.privacyConsents || []);

    // Party Selection State
    const [partyId, setPartyId] = useState(customer.partyId);
    const [partyType, setPartyType] = useState<string | undefined>(customer.partyType);
    const [partyName, setPartyName] = useState<string | undefined>(customer.partyName);

    const [relatedParties, setRelatedParties] = useState<Partial<RelatedParty>[]>(customer.relatedParties || []);

    // Removed local useParties logic in favor of PartySelector component

    // Removed selectedPartyFromSearch derivation logic since we store values explicitly

    // New State
    // Customer has creditProfiles array, but UI treats as single managed profile for now (taking first or new)
    const [creditProfile, setCreditProfile] = useState<Partial<CreditProfile> | null>(customer.creditProfiles?.[0] || null);
    const [accounts, setAccounts] = useState<Partial<CustomerAccount>[]>(customer.accounts || []);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        try {
            await updateMutation.mutateAsync({
                id: customer.id,
                status,
                name,
                partyId: partyId !== customer.partyId ? partyId : undefined,


                privacyConsents: privacyConsents as PrivacyConsent[],
                accounts: accounts as CustomerAccount[],
                creditProfiles: creditProfile ? [creditProfile as CreditProfile] : [],
                relatedParties: relatedParties as RelatedParty[],
            });
            navigate(`/customers/${customer.id}`);
        } catch (err) {
            console.error('Failed to update customer:', err);
        }
    };



    const addPrivacyConsent = () => {
        setPrivacyConsents((prev) => [
            ...prev,
            { consentType: '', status: 'pending' },
        ]);
    };

    const removePrivacyConsent = (index: number) => {
        setPrivacyConsents((prev) => prev.filter((_, i) => i !== index));
    };

    // Credit Profile Handlers
    const toggleCreditProfile = () => {
        if (creditProfile) {
            setCreditProfile(null);
        } else {
            setCreditProfile({ creditRiskScore: 0, creditScore: 0 });
        }
    };

    // Account Handlers
    const addAccount = () => {
        setAccounts(prev => [...prev, { name: '', accountType: '', accountStatus: 'active' }]);
    };
    const removeAccount = (index: number) => {
        setAccounts(prev => prev.filter((_, i) => i !== index));
    };

    return (
        <>
            <div className="page-header">
                <div className="page-header-content mb-3">
                    <h2>Edit Customer</h2>
                    <p className="lead text-muted mb-0">Editing {customer.name}</p>
                </div>
                <button
                    type="submit"
                    form="edit-form"
                    className="btn btn-primary"
                    disabled={updateMutation.isPending}
                >
                    {updateMutation.isPending ? <Loader2 className="spin" size={18} /> : <Save size={18} />}
                    <span>Save Changes</span>
                </button>
            </div>

            <form id="edit-form" className="form-container" onSubmit={handleSubmit}>
                {/* Linked Party */}
                <div className="card form-section mt-4">
                    <h3>Linked Party</h3>
                    <p className="section-description">
                        The Individual or Organization linked to this customer account.
                    </p>

                    <PartyPicker
                        value={partyId ? {
                            id: partyId,
                            name: partyName,
                            '@type': partyType
                        } : null}
                        onChange={(party) => {
                            if (party) {
                                setPartyId(party.id);
                                setPartyType(party['@type']);
                                setPartyName(getPartyDisplayName(party));
                            }
                        }}
                        customActions={partyId !== customer.partyId ? (
                            <button
                                type="button"
                                className="btn btn-outline-danger btn-sm"
                                onClick={() => {
                                    setPartyId(customer.partyId);
                                    setPartyType(customer.partyType);
                                    setPartyName(customer.partyName);
                                }}
                            >
                                Revert
                            </button>
                        ) : undefined}
                    />
                </div>

                {/* Related Parties */}
                <RelatedPartiesForm
                    items={relatedParties}
                    onChange={setRelatedParties}
                />

                {/* Basic Info */}
                <div className="card form-section">
                    <h3>Basic Information</h3>
                    <div className="form-grid">
                        <div className="form-group">
                            <label htmlFor="name">Customer Name</label>
                            <input
                                id="name"
                                type="text"
                                value={name}
                                onChange={(e) => setName(e.target.value)}
                                required
                            />
                        </div>
                        <div className="form-group">
                            <label htmlFor="status">Status</label>
                            <select
                                id="status"
                                value={status}
                                onChange={(e) => setStatus(e.target.value as CustomerStatus)}
                            >
                                <option value="prospecting">Prospecting</option>
                                <option value="active">Active</option>
                                <option value="suspended">Suspended</option>
                                <option value="inactive">Inactive</option>
                            </select>
                        </div>
                    </div>
                </div>

                {/* Credit Profile */}
                <div className="card form-section">
                    <div className="section-header">
                        <h3>Credit Profile</h3>
                        <button
                            type="button"
                            className={`btn btn-sm ${creditProfile ? 'btn-danger-outline' : 'btn-secondary'}`}
                            onClick={toggleCreditProfile}
                        >
                            {creditProfile ? <Trash2 size={16} /> : <Plus size={16} />}
                            <span>{creditProfile ? 'Remove Profile' : 'Add Profile'}</span>
                        </button>
                    </div>

                    {creditProfile && (
                        <div className="form-grid">
                            <div className="form-group">
                                <label htmlFor="credit-risk">Credit Risk Score</label>
                                <input
                                    id="credit-risk"
                                    type="number"
                                    value={creditProfile.creditRiskScore || 0}
                                    onChange={(e) => setCreditProfile({ ...creditProfile, creditRiskScore: parseInt(e.target.value) || 0 })}
                                />
                            </div>
                            <div className="form-group">
                                <label htmlFor="credit-score">Credit Score</label>
                                <input
                                    id="credit-score"
                                    type="number"
                                    value={creditProfile.creditScore || 0}
                                    onChange={(e) => setCreditProfile({ ...creditProfile, creditScore: parseInt(e.target.value) || 0 })}
                                />
                            </div>
                            <div className="form-group">
                                <label htmlFor="credit-valid-from">Valid From</label>
                                <input
                                    id="credit-valid-from"
                                    type="date"
                                    value={creditProfile.validForStart?.split('T')[0] || ''}
                                    onChange={(e) => setCreditProfile({ ...creditProfile, validForStart: e.target.value })}
                                />
                            </div>
                        </div>
                    )}
                </div>

                {/* Customer Accounts */}
                <div className="card form-section">
                    <div className="section-header">
                        <h3>Accounts</h3>
                        <button type="button" className="btn btn-secondary btn-sm" onClick={addAccount}>
                            <Plus size={16} />
                            <span>Add Account</span>
                        </button>
                    </div>

                    {accounts.length === 0 ? (
                        <p className="empty-text">No accounts added</p>
                    ) : (
                        <div className="repeatable-list">
                            {accounts.map((account, index) => (
                                <div key={index} className="repeatable-item">
                                    <div className="form-grid" style={{ gridTemplateColumns: 'repeat(3, 1fr) auto', gap: '1rem' }}>
                                        <div className="form-group">
                                            <label htmlFor={`acc-name-${index}`}>Account Name</label>
                                            <input
                                                id={`acc-name-${index}`}
                                                type="text"
                                                value={account.name}
                                                onChange={(e) => {
                                                    const updated = [...accounts];
                                                    updated[index] = { ...account, name: e.target.value };
                                                    setAccounts(updated);
                                                }}
                                                required
                                            />
                                        </div>
                                        <div className="form-group">
                                            <label htmlFor={`acc-type-${index}`}>Type</label>
                                            <input
                                                id={`acc-type-${index}`}
                                                type="text"
                                                value={account.accountType}
                                                onChange={(e) => {
                                                    const updated = [...accounts];
                                                    updated[index] = { ...account, accountType: e.target.value };
                                                    setAccounts(updated);
                                                }}
                                                placeholder="e.g. Savings, Checking"
                                            />
                                        </div>
                                        <div className="form-group">
                                            <label htmlFor={`acc-status-${index}`}>Status</label>
                                            <select
                                                id={`acc-status-${index}`}
                                                value={account.accountStatus}
                                                onChange={(e) => {
                                                    const updated = [...accounts];
                                                    updated[index] = { ...account, accountStatus: e.target.value as 'active' | 'inactive' | 'suspended' };
                                                    setAccounts(updated);
                                                }}
                                            >
                                                <option value="active">Active</option>
                                                <option value="inactive">Inactive</option>
                                                <option value="suspended">Suspended</option>
                                            </select>
                                        </div>
                                        <div className="form-group form-group--action">
                                            <button type="button" className="btn-icon btn-icon--danger" onClick={() => removeAccount(index)}>
                                                <Trash2 size={16} />
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>



                {/* Privacy Consents */}
                <div className="card form-section">
                    <div className="section-header">
                        <h3>Privacy Consents</h3>
                        <button type="button" className="btn btn-secondary btn-sm" onClick={addPrivacyConsent}>
                            <Plus size={16} />
                            <span>Add Consent</span>
                        </button>
                    </div>

                    {privacyConsents.length === 0 ? (
                        <p className="empty-text">No privacy consents</p>
                    ) : (
                        <div className="repeatable-list">
                            {privacyConsents.map((consent, index) => (
                                <div key={index} className="repeatable-item">
                                    <div className="form-grid" style={{ gridTemplateColumns: 'repeat(2, 1fr) auto', gap: '1rem' }}>
                                        <div className="form-group">
                                            <label>Consent Type</label>
                                            <input
                                                type="text"
                                                value={consent.consentType || ''}
                                                onChange={(e) => {
                                                    const updated = [...privacyConsents];
                                                    updated[index] = { ...updated[index], consentType: e.target.value };
                                                    setPrivacyConsents(updated);
                                                }}
                                                placeholder="e.g., Marketing, Analytics"
                                            />
                                        </div>
                                        <div className="form-group">
                                            <label>Status</label>
                                            <select
                                                value={consent.status || 'pending'}
                                                onChange={(e) => {
                                                    const updated = [...privacyConsents];
                                                    updated[index] = { ...updated[index], status: e.target.value as 'given' | 'revoked' | 'pending' };
                                                    setPrivacyConsents(updated);
                                                }}
                                            >
                                                <option value="pending">Pending</option>
                                                <option value="given">Given</option>
                                                <option value="revoked">Revoked</option>
                                            </select>
                                        </div>
                                        <div className="form-group form-group--action">
                                            <button
                                                type="button"
                                                className="btn-icon btn-icon--danger"
                                                onClick={() => removePrivacyConsent(index)}
                                            >
                                                <Trash2 size={16} />
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            </form>
        </>
    );
}
