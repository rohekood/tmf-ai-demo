import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { ArrowLeft, UserPlus, Loader2, Plus, Trash2 } from 'lucide-react';
import { useOnboardCustomer } from './api';
import PartySelector from '../parties/PartySelector';
import { getPartyDisplayName } from '../parties/types';
import type { OnboardCustomerPayload, CreditProfile, CustomerAccount, TaxExemption } from './types';
import type { PartyUnion } from '../parties/types';
import '../parties/PartyFormPage.css';
import './CustomerOnboardPage.css';

export default function CustomerOnboardPage() {
    const navigate = useNavigate();
    const onboardMutation = useOnboardCustomer();

    const [selectedParty, setSelectedParty] = useState<PartyUnion | null>(null as PartyUnion | null);
    const [customerName, setCustomerName] = useState('');

    const [privacyConsents, setPrivacyConsents] = useState<{ consentType: string; status: 'given' | 'revoked' | 'pending' }[]>([]);

    // New State for enhancements
    const [creditProfile, setCreditProfile] = useState<Omit<CreditProfile, 'id'> | null>(null);
    const [accounts, setAccounts] = useState<Omit<CustomerAccount, 'id'>[]>([]);
    const [taxExemptions, setTaxExemptions] = useState<Omit<TaxExemption, 'id'>[]>([]);

    // Prefill customer name when party is selected
    useEffect(() => {
        if (selectedParty) {
            // Only update if name is empty
            if (!customerName) {
                setCustomerName(getPartyDisplayName(selectedParty));
            }
        }
    }, [selectedParty]);



    // ... handlers ...
    // (Omitting handlers for brevity as they are unchanged usually, but I need to make sure I don't break the file structure.
    //  The previous ReplaceFileContent targeted lines 132-200. This is line 43-69.)
    // Wait, I can only do ONE block. I'll split this.

    // Update handleSubmit
    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!selectedParty || !customerName) {
            alert('Please select a party and enter a customer name');
            return;
        }

        const payload: OnboardCustomerPayload = {
            name: customerName,
            partyId: selectedParty.id,
            privacyConsents: privacyConsents,
            creditProfiles: creditProfile ? [{
                creditRiskScore: creditProfile.creditRiskScore,
                creditScore: creditProfile.creditScore,
                validForStart: creditProfile.validForStart
            }] : [],
            accounts: accounts,
            taxExemptions: taxExemptions,
        };

        try {
            await onboardMutation.mutateAsync(payload);
            navigate('/customers');
        } catch (err) {
            console.error('Failed to onboard customer:', err);
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

    // Tax Exemption Handlers
    const addTaxExemption = () => {
        setTaxExemptions(prev => [...prev, { certificateNumber: '', issuingJurisdiction: '' }]);
    };
    const removeTaxExemption = (index: number) => {
        setTaxExemptions(prev => prev.filter((_, i) => i !== index));
    };

    return (
        <div className="customer-onboard-page">
            <div className="page-header">
                <div className="page-header-content">
                    <Link to="/customers" className="back-link">
                        <ArrowLeft size={18} />
                        <span>Back to Customers</span>
                    </Link>
                    <h2>Onboard Customer</h2>
                    <p className="page-description">Create a new customer linked to an existing party</p>
                </div>
                <button
                    type="submit"
                    form="onboard-form"
                    className="btn btn-primary"
                    disabled={onboardMutation.isPending || !selectedParty || !customerName}
                >
                    {onboardMutation.isPending ? <Loader2 className="spin" size={18} /> : <UserPlus size={18} />}
                    <span>Onboard Customer</span>
                </button>
            </div>

            <form id="onboard-form" className="form-container" onSubmit={handleSubmit}>
                {/* Party Selection */}
                <div className="card form-section">
                    <h3>Select Party</h3>
                    <p className="section-description">
                        Search and select an existing Individual or Organization to link to this customer.
                    </p>

                    {selectedParty ? (
                        <div className="selected-party">
                            <div className="selected-party-info">
                                <span className={`party-type-badge ${selectedParty['@type'].toLowerCase()}`}>
                                    {selectedParty['@type']}
                                </span>
                                <span className="selected-party-name">{getPartyDisplayName(selectedParty)}</span>
                            </div>
                            <button
                                type="button"
                                className="btn btn-secondary btn-sm"
                                onClick={() => setSelectedParty(null)}
                            >
                                Change
                            </button>
                        </div>
                    ) : (
                        <PartySelector
                            selectedPartyId={(selectedParty as any)?.id}
                            onSelect={(party) => setSelectedParty(party)}
                        />
                    )}
                </div>

                {/* Customer Details */}
                <div className="card form-section">
                    <h3>Customer Details</h3>

                    <div className="form-grid">
                        <div className="form-group form-group--full">
                            <label htmlFor="customerName">Customer Name *</label>
                            <input
                                id="customerName"
                                type="text"
                                value={customerName}
                                onChange={(e) => setCustomerName(e.target.value)}
                                required
                                placeholder="Enter customer display name"
                            />
                            <span className="form-hint">
                                This name will be used to identify the customer relationship
                            </span>
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
                                <label htmlFor="credit-risk-score">Credit Risk Score</label>
                                <input
                                    id="credit-risk-score"
                                    type="number"
                                    value={creditProfile.creditRiskScore}
                                    onChange={(e) => setCreditProfile({ ...creditProfile, creditRiskScore: parseInt(e.target.value) || 0 })}
                                />
                            </div>
                            <div className="form-group">
                                <label htmlFor="credit-score">Credit Score</label>
                                <input
                                    id="credit-score"
                                    type="number"
                                    value={creditProfile.creditScore}
                                    onChange={(e) => setCreditProfile({ ...creditProfile, creditScore: parseInt(e.target.value) || 0 })}
                                />
                            </div>
                            <div className="form-group">
                                <label htmlFor="valid-for-start">Valid From</label>
                                <input
                                    id="valid-for-start"
                                    type="date"
                                    value={creditProfile.validForStart || ''}
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
                                    <div className="form-grid">
                                        <div className="form-group">
                                            <label htmlFor={`account-name-${index}`}>Account Name</label>
                                            <input
                                                id={`account-name-${index}`}
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
                                            <label htmlFor={`account-type-${index}`}>Type</label>
                                            <input
                                                id={`account-type-${index}`}
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
                                            <label htmlFor={`account-status-${index}`}>Status</label>
                                            <select
                                                id={`account-status-${index}`}
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

                {/* Tax Exemptions */}
                <div className="card form-section">
                    <div className="section-header">
                        <h3>Tax Exemptions</h3>
                        <button type="button" className="btn btn-secondary btn-sm" onClick={addTaxExemption}>
                            <Plus size={16} />
                            <span>Add Exemption</span>
                        </button>
                    </div>

                    {taxExemptions.length === 0 ? (
                        <p className="empty-text">No tax exemptions added</p>
                    ) : (
                        <div className="repeatable-list">
                            {taxExemptions.map((exemption, index) => (
                                <div key={index} className="repeatable-item">
                                    <div className="form-grid">
                                        <div className="form-group">
                                            <label htmlFor={`tax-cert-${index}`}>Certificate Number</label>
                                            <input
                                                id={`tax-cert-${index}`}
                                                type="text"
                                                value={exemption.certificateNumber}
                                                onChange={(e) => {
                                                    const updated = [...taxExemptions];
                                                    updated[index] = { ...exemption, certificateNumber: e.target.value };
                                                    setTaxExemptions(updated);
                                                }}
                                                required
                                            />
                                        </div>
                                        <div className="form-group">
                                            <label htmlFor={`tax-jur-${index}`}>Issuing Jurisdiction</label>
                                            <input
                                                id={`tax-jur-${index}`}
                                                type="text"
                                                value={exemption.issuingJurisdiction}
                                                onChange={(e) => {
                                                    const updated = [...taxExemptions];
                                                    updated[index] = { ...exemption, issuingJurisdiction: e.target.value };
                                                    setTaxExemptions(updated);
                                                }}
                                                required
                                            />
                                        </div>
                                        <div className="form-group form-group--action">
                                            <button type="button" className="btn-icon btn-icon--danger" onClick={() => removeTaxExemption(index)}>
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
                                    <div className="form-grid">
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
                                                required
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
        </div>
    );
}
