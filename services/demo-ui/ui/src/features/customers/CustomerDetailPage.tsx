
import { useState } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { ArrowLeft, Edit, Trash2, CreditCard, Shield, Loader2, ExternalLink, MessageSquarePlus } from 'lucide-react';
import LogInteractionModal from './components/LogInteractionModal';
import RelatedPartiesList from './components/RelatedPartiesList';
import PaymentMethodsList from './components/PaymentMethodsList';
import MarketSegmentsList from './components/MarketSegmentsList';

import InteractionsList from './components/InteractionsList';
import { useCustomer, useDeleteCustomer } from './api';
import '../parties/PartyDetailPage.css';
import './CustomerDetailPage.css';

export default function CustomerDetailPage() {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const { data: customer, isLoading, error } = useCustomer(id);
    const deleteMutation = useDeleteCustomer();
    const [showLogInteraction, setShowLogInteraction] = useState(false);

    const handleDelete = () => {
        if (customer && confirm(`Are you sure you want to delete customer "${customer.name}" ? `)) {
            deleteMutation.mutate(customer.id, {
                onSuccess: () => navigate('/customers'),
            });
        }
    };

    if (isLoading) {
        return (
            <div className="customer-detail-page">
                <div className="loading-state" role="status">
                    <Loader2 className="spin" size={32} />
                    <p>Loading customer details...</p>
                </div>
            </div>
        );
    }

    if (error || !customer) {
        return (
            <div className="customer-detail-page">
                <div className="error-state card" role="alert">
                    <p>Failed to load customer: {error?.message || 'Customer not found'}</p>
                    <Link to="/customers" className="btn btn-secondary">
                        Back to Customers
                    </Link>
                </div>
            </div>
        );
    }

    return (
        <div className="customer-detail-page">
            <div className="page-header">
                <div className="page-header-content">
                    <Link to="/customers" className="back-link">
                        <ArrowLeft size={18} />
                        <span>Back to Customers</span>
                    </Link>
                    <h2>{customer.name}</h2>
                    <div className="customer-meta">
                        <span className={`status - badge ${customer.status} `}>
                            {customer.status}
                        </span>
                        <Link to={`/parties/${customer.partyId}`} className="party-ref">
                            Party: {customer.partyName || customer.partyId.slice(0, 8)}
                            <ExternalLink size={14} />
                        </Link>
                    </div>
                </div>
                <div className="page-actions">
                    <button
                        className="btn btn-secondary"
                        onClick={() => setShowLogInteraction(true)}
                    >
                        <MessageSquarePlus size={18} />
                        <span>Log Interaction</span>
                    </button>
                    <Link to={`/customers/${id}/edit`} className="btn btn-secondary" >
                        <Edit size={18} />
                        <span>Edit</span>
                    </Link >
                    <button
                        className="btn btn-danger"
                        onClick={handleDelete}
                        disabled={deleteMutation.isPending}
                    >
                        <Trash2 size={18} />
                        <span>Delete</span>
                    </button>
                </div >
            </div >

            <div className="detail-grid">
                {/* Basic Info */}
                <div className="card detail-card">
                    <h3>Basic Information</h3>
                    <dl className="detail-list">
                        <div className="detail-item">
                            <dt>Customer ID</dt>
                            <dd className="mono">{customer.id}</dd>
                        </div>
                        <div className="detail-item">
                            <dt>Name</dt>
                            <dd>{customer.name}</dd>
                        </div>
                        <div className="detail-item">
                            <dt>Status</dt>
                            <dd>
                                <span className={`status-badge ${customer.status}`}>{customer.status}</span>
                            </dd>
                        </div>
                        <div className="detail-item">
                            <dt>Party Reference</dt>
                            <dd>
                                <Link to={`/parties/${customer.partyId}`} className="party-link">
                                    {customer.partyId}
                                    <ExternalLink size={12} />
                                </Link>
                            </dd>
                        </div>
                    </dl>
                </div>

                {/* Credit Profile */}
                <div className="card detail-card">
                    <h3>
                        <CreditCard size={18} />
                        <span>Credit Profile</span>
                    </h3>
                    {customer.creditProfiles && customer.creditProfiles.length > 0 ? (
                        <dl className="detail-list">
                            <div className="detail-item">
                                <dt>Credit Score</dt>
                                <dd className="score-value">{customer.creditProfiles[0].creditScore}</dd>
                            </div>
                            <div className="detail-item">
                                <dt>Risk Score</dt>
                                <dd className="score-value">{customer.creditProfiles[0].creditRiskScore}</dd>
                            </div>
                        </dl>
                    ) : (
                        <p className="empty-text">No credit profile</p>
                    )}
                </div>

                {/* Customer Accounts */}
                <div className="card detail-card">
                    <h3>
                        <span>Customer Accounts</span>
                        <span className="count-badge">{customer.accounts?.length || 0}</span>
                    </h3>
                    {customer.accounts && customer.accounts.length > 0 ? (
                        <ul className="account-list">
                            {customer.accounts.map((account) => (
                                <li key={account.id} className="account-item">
                                    <div className="account-content">
                                        <span className="account-name">{account.name}</span>
                                        <span className="account-type">{account.accountType}</span>
                                    </div>
                                    <span className={`status-badge ${account.accountStatus}`}>
                                        {account.accountStatus}
                                    </span>
                                </li>
                            ))}
                        </ul>
                    ) : (
                        <p className="empty-text">No customer accounts</p>
                    )}
                </div>

                {/* Privacy Consents */}
                <div className="card detail-card">
                    <h3>
                        <Shield size={18} />
                        <span>Privacy Consents</span>
                        <span className="count-badge">{customer.privacyConsents?.length || 0}</span>
                    </h3>
                    {customer.privacyConsents && customer.privacyConsents.length > 0 ? (
                        <ul className="consent-list">
                            {customer.privacyConsents.map((consent) => (
                                <li key={consent.id} className="consent-item">
                                    <div className="consent-content">
                                        <span className="consent-type">{consent.consentType}</span>
                                    </div>
                                    <span className={`consent-status consent-status--${consent.status}`}>
                                        {consent.status}
                                    </span>
                                </li>
                            ))}
                        </ul>
                    ) : (
                        <p className="empty-text">No privacy consents</p>
                    )}
                </div>



                {/* Related Parties */}
                <div className="card detail-card">
                    <h3>Related Parties</h3>
                    <RelatedPartiesList items={customer.relatedParties || []} />
                </div>

                {/* Payment Methods */}
                <div className="card detail-card">
                    <h3>Payment Methods</h3>
                    <PaymentMethodsList items={customer.paymentMethods || []} />
                </div>

                {/* Market Segments */}
                <div className="card detail-card">
                    <h3>Market Segments</h3>
                    <MarketSegmentsList items={customer.marketSegments || []} />
                </div>



                {/* Interactions */}
                <div className="card detail-card">
                    <h3>Interactions</h3>
                    <InteractionsList items={customer.customerInteractions || []} />
                </div>
            </div>

            {
                showLogInteraction && customer && (
                    <LogInteractionModal
                        customerId={customer.id}
                        onClose={() => setShowLogInteraction(false)}
                    />
                )
            }
        </div >
    );
}
