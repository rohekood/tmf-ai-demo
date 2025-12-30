import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { ArrowLeft, UserPlus, Loader2 } from 'lucide-react';
import { useOnboardCustomer } from './api';
import PartySelector from '../parties/PartySelector';
import { getPartyDisplayName } from '../parties/types';
import type { PartyUnion } from '../parties/types';
import type {
    OnboardCustomerPayload,
    PrivacyConsent,
    CreditProfile,
    CustomerAccount,
    RelatedParty,
    PaymentMethod,
    MarketSegment,
} from './types';
import RelatedPartiesForm from './components/RelatedPartiesForm';
import PaymentMethodsForm from './components/PaymentMethodsForm';
import MarketSegmentsForm from './components/MarketSegmentsForm';

export default function CustomerOnboardPage() {
    const navigate = useNavigate();
    const onboardMutation = useOnboardCustomer();

    const [selectedParty, setSelectedParty] = useState<PartyUnion | null>(null);
    const [customerName, setCustomerName] = useState('');
    const [privacyConsents, setPrivacyConsents] = useState<Omit<PrivacyConsent, 'id'>[]>([]);
    const [creditProfile, setCreditProfile] = useState<Omit<CreditProfile, 'id'> | null>(null);
    const [accounts, setAccounts] = useState<Omit<CustomerAccount, 'id'>[]>([]);
    // New state for TMF629 features
    const [relatedParties, setRelatedParties] = useState<Omit<RelatedParty, 'id'>[]>([]);
    const [paymentMethods, setPaymentMethods] = useState<Omit<PaymentMethod, 'id'>[]>([]);
    const [marketSegments, setMarketSegments] = useState<Omit<MarketSegment, 'id'>[]>([]);




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
            relatedParties: relatedParties,
            paymentMethods: paymentMethods,
            marketSegments: marketSegments,

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
        setAccounts(prev => [...prev, { name: '', accountType: '', accountStatus: 'active', billFormat: 'PDF', billingCycle: 'Monthly' }]);
    };
    const removeAccount = (index: number) => {
        setAccounts(prev => prev.filter((_, i) => i !== index));
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
                            onSelect={(party) => {
                                setSelectedParty(party);
                                // Prefill name if empty or if needed (here we force prefill on selection)
                                if (!customerName) {
                                    setCustomerName(getPartyDisplayName(party));
                                }
                            }}
                        />
                    )}
                </div>

                {/* Customer Details */}
                <div className="card form-section">
                    <h3>Customer Details</h3>
                    <div className="form-group">
                        <label htmlFor="customerName">Customer Name</label>
                        <input
                            id="customerName"
                            type="text"
                            value={customerName}
                            onChange={(e) => setCustomerName(e.target.value)}
                            placeholder="e.g. John Doe or Acme Corp"
                            required
                        />
                    </div>
                </div>

                {/* Privacy Consents */}
                <div className="card form-section">
                    <div className="section-header">
                        <h3>Privacy Consents</h3>
                        <button type="button" className="btn btn-secondary btn-sm" onClick={addPrivacyConsent}>
                            Add Consent
                        </button>
                    </div>
                    {privacyConsents.length > 0 ? (
                        <div className="items-list">
                            {privacyConsents.map((consent, index) => (
                                <div key={index} className="item-row">
                                    <div className="form-group">
                                        <input
                                            type="text"
                                            value={consent.consentType}
                                            onChange={(e) => {
                                                const updated = [...privacyConsents];
                                                updated[index].consentType = e.target.value;
                                                setPrivacyConsents(updated);
                                            }}
                                            placeholder="Consent Type"
                                        />
                                    </div>
                                    <button
                                        type="button"
                                        className="btn-icon danger"
                                        onClick={() => removePrivacyConsent(index)}
                                    >
                                        &times;
                                    </button>
                                </div>
                            ))}
                        </div>
                    ) : (
                        <p className="empty-text">No privacy consents added</p>
                    )}
                </div>

                {/* Credit Profile */}
                <div className="card form-section">
                    <div className="section-header">
                        <h3>Credit Profile</h3>
                        <div className="form-check">
                            <input
                                type="checkbox"
                                id="hasCreditProfile"
                                checked={!!creditProfile}
                                onChange={toggleCreditProfile}
                            />
                            <label htmlFor="hasCreditProfile">Add Credit Profile</label>
                        </div>
                    </div>

                    {creditProfile && (
                        <div className="form-grid">
                            <div className="form-group">
                                <label htmlFor="creditScore">Credit Score</label>
                                <input
                                    id="creditScore"
                                    type="number"
                                    value={creditProfile.creditScore || ''}
                                    onChange={(e) => setCreditProfile({ ...creditProfile, creditScore: parseInt(e.target.value) || 0 })}
                                />
                            </div>
                            <div className="form-group">
                                <label htmlFor="riskScore">Risk Score</label>
                                <input
                                    id="riskScore"
                                    type="number"
                                    value={creditProfile.creditRiskScore || ''}
                                    onChange={(e) => setCreditProfile({ ...creditProfile, creditRiskScore: parseInt(e.target.value) || 0 })}
                                />
                            </div>
                        </div>
                    )}
                </div>

                {/* Customer Accounts */}
                <div className="card form-section">
                    <div className="section-header">
                        <h3>Customer Accounts</h3>
                        <button type="button" className="btn btn-secondary btn-sm" onClick={addAccount}>
                            Add Account
                        </button>
                    </div>
                    {accounts.length > 0 ? (
                        <div className="items-list">
                            {accounts.map((account, index) => (
                                <div key={index} className="item-row">
                                    <div className="form-group">
                                        <input
                                            type="text"
                                            value={account.name}
                                            onChange={(e) => {
                                                const updated = [...accounts];
                                                updated[index].name = e.target.value;
                                                setAccounts(updated);
                                            }}
                                            placeholder="Account Name"
                                        />
                                    </div>
                                    <div className="form-group">
                                        <input
                                            type="text"
                                            value={account.accountType}
                                            onChange={(e) => {
                                                const updated = [...accounts];
                                                updated[index].accountType = e.target.value;
                                                setAccounts(updated);
                                            }}
                                            placeholder="Type"
                                        />
                                    </div>
                                    <div className="form-group">
                                        <input
                                            type="text"
                                            value={account.billFormat}
                                            onChange={(e) => {
                                                const updated = [...accounts];
                                                updated[index].billFormat = e.target.value;
                                                setAccounts(updated);
                                            }}
                                            placeholder="Bill Format"
                                        />
                                    </div>
                                    <div className="form-group">
                                        <input
                                            type="text"
                                            value={account.billingCycle}
                                            onChange={(e) => {
                                                const updated = [...accounts];
                                                updated[index].billingCycle = e.target.value;
                                                setAccounts(updated);
                                            }}
                                            placeholder="Cycle"
                                        />
                                    </div>
                                    <button
                                        type="button"
                                        className="btn-icon danger"
                                        onClick={() => removeAccount(index)}
                                    >
                                        &times;
                                    </button>
                                </div>
                            ))}
                        </div>
                    ) : (
                        <p className="empty-text">No accounts added</p>
                    )}
                </div>

                {/* Related Parties - New Component */}
                <RelatedPartiesForm items={relatedParties} onChange={setRelatedParties} />

                {/* Payment Methods - New Component */}
                <PaymentMethodsForm items={paymentMethods} onChange={setPaymentMethods} />

                {/* Market Segments - New Component */}
                <MarketSegmentsForm items={marketSegments} onChange={setMarketSegments} />



            </form>
        </div>
    );
}
