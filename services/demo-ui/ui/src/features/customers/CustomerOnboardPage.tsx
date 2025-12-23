import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { ArrowLeft, UserPlus, Search, Check, Loader2 } from 'lucide-react';
import { useOnboardCustomer } from './api';
import { useParties } from '../parties/api';
import { getPartyDisplayName } from '../parties/types';
import type { OnboardCustomerPayload } from './types';
import '../parties/PartyFormPage.css';
import './CustomerOnboardPage.css';

export default function CustomerOnboardPage() {
    const navigate = useNavigate();
    const onboardMutation = useOnboardCustomer();

    const [partySearch, setPartySearch] = useState('');
    const [selectedPartyId, setSelectedPartyId] = useState<string | null>(null);
    const [customerName, setCustomerName] = useState('');

    const { data: parties = [], isLoading: loadingParties } = useParties(
        partySearch ? { givenName: partySearch, tradingName: partySearch } : undefined
    );

    const selectedParty = parties.find((p) => p.id === selectedPartyId);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!selectedPartyId || !customerName) {
            alert('Please select a party and enter a customer name');
            return;
        }

        const payload: OnboardCustomerPayload = {
            name: customerName,
            partyId: selectedPartyId,
        };

        try {
            await onboardMutation.mutateAsync(payload);
            navigate('/customers');
        } catch (err) {
            console.error('Failed to onboard customer:', err);
        }
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
                    disabled={onboardMutation.isPending || !selectedPartyId || !customerName}
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

                    <div className="party-search">
                        <Search size={18} className="search-icon" />
                        <input
                            type="text"
                            placeholder="Search parties by name..."
                            value={partySearch}
                            onChange={(e) => setPartySearch(e.target.value)}
                            className="search-input-full"
                        />
                    </div>

                    {selectedParty && (
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
                                onClick={() => setSelectedPartyId(null)}
                            >
                                Change
                            </button>
                        </div>
                    )}

                    {!selectedParty && (
                        <div className="party-results">
                            {loadingParties ? (
                                <div className="loading-inline">
                                    <Loader2 className="spin" size={16} />
                                    <span>Searching...</span>
                                </div>
                            ) : parties.length === 0 ? (
                                <p className="empty-text">
                                    {partySearch ? 'No parties found. Try a different search term.' : 'Enter a name to search for parties.'}
                                </p>
                            ) : (
                                <>
                                    {parties.map((party) => (
                                        <div
                                            key={party.id}
                                            className={`party-item ${selectedPartyId === party.id ? 'selected' : ''}`}
                                            onClick={() => setSelectedPartyId(party.id)}
                                            role="button"
                                            tabIndex={0}
                                        >
                                            <div className="party-item-info">
                                                <span className={`party-type-badge ${party['@type'].toLowerCase()}`}>
                                                    {party['@type']}
                                                </span>
                                                <span className="party-item-name">{getPartyDisplayName(party)}</span>
                                            </div>
                                            {selectedPartyId === party.id && <Check size={18} className="check-icon" />}
                                        </div>
                                    ))}
                                </>
                            )}
                        </div>
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
            </form>
        </div>
    );
}
