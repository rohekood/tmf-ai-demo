import { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { ArrowLeft, Save, Plus, Trash2, Loader2 } from 'lucide-react';
import { useCustomer, useUpdateCustomer } from './api';
import type { CustomerStatus, TaxExemption, PrivacyConsent } from './types';
import '../parties/PartyFormPage.css';

export default function CustomerEditPage() {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const { data: customer, isLoading } = useCustomer(id);
    const updateMutation = useUpdateCustomer();

    const [status, setStatus] = useState<CustomerStatus>('active');
    const [name, setName] = useState('');
    const [taxExemptions, setTaxExemptions] = useState<Partial<TaxExemption>[]>([]);
    const [privacyConsents, setPrivacyConsents] = useState<Partial<PrivacyConsent>[]>([]);

    useEffect(() => {
        if (customer) {
            setStatus(customer.status);
            setName(customer.name);
            setTaxExemptions(customer.taxExemptions || []);
            setPrivacyConsents(customer.privacyConsents || []);
        }
    }, [customer]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!id) return;

        try {
            await updateMutation.mutateAsync({
                id,
                status,
                name,
                taxExemptions: taxExemptions as TaxExemption[],
                privacyConsents: privacyConsents as PrivacyConsent[],
            });
            navigate(`/customers/${id}`);
        } catch (err) {
            console.error('Failed to update customer:', err);
        }
    };

    const addTaxExemption = () => {
        setTaxExemptions((prev) => [
            ...prev,
            { certificateNumber: '', issuingJurisdiction: '' },
        ]);
    };

    const removeTaxExemption = (index: number) => {
        setTaxExemptions((prev) => prev.filter((_, i) => i !== index));
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

    if (isLoading) {
        return (
            <div className="customer-edit-page">
                <div className="loading-state">
                    <Loader2 className="spin" size={32} />
                    <p>Loading customer...</p>
                </div>
            </div>
        );
    }

    return (
        <div className="customer-edit-page">
            <div className="page-header">
                <div className="page-header-content">
                    <Link to={`/customers/${id}`} className="back-link">
                        <ArrowLeft size={18} />
                        <span>Back to Details</span>
                    </Link>
                    <h2>Edit Customer</h2>
                    <p className="page-description">Editing {customer?.name}</p>
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
                        <p className="empty-text">No tax exemptions</p>
                    ) : (
                        <div className="repeatable-list">
                            {taxExemptions.map((exemption, index) => (
                                <div key={index} className="repeatable-item">
                                    <div className="form-grid">
                                        <div className="form-group">
                                            <label>Certificate Number</label>
                                            <input
                                                type="text"
                                                value={exemption.certificateNumber || ''}
                                                onChange={(e) => {
                                                    const updated = [...taxExemptions];
                                                    updated[index] = { ...updated[index], certificateNumber: e.target.value };
                                                    setTaxExemptions(updated);
                                                }}
                                                placeholder="Certificate #"
                                            />
                                        </div>
                                        <div className="form-group">
                                            <label>Jurisdiction</label>
                                            <input
                                                type="text"
                                                value={exemption.issuingJurisdiction || ''}
                                                onChange={(e) => {
                                                    const updated = [...taxExemptions];
                                                    updated[index] = { ...updated[index], issuingJurisdiction: e.target.value };
                                                    setTaxExemptions(updated);
                                                }}
                                                placeholder="e.g., US-CA"
                                            />
                                        </div>
                                        <div className="form-group form-group--action">
                                            <button
                                                type="button"
                                                className="btn-icon btn-icon--danger"
                                                onClick={() => removeTaxExemption(index)}
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
