import { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { ArrowLeft, Save, Plus, Trash2, Loader2 } from 'lucide-react';
import { useParty, useCreateParty, useUpdateParty } from './api';
import type { PartyType, CreatePartyPayload, UpdatePartyPayload, ContactMedium, Identification, TaxExemption, ExternalReference, Attachment, RelatedParty } from './types';
import PartySelector from './PartySelector';
import './PartyFormPage.css';

interface FormState {
    type: PartyType;
    // Individual fields
    givenName: string;
    familyName: string;
    middleName: string;
    birthDate: string;
    gender: string;
    // Organization fields
    tradingName: string;
    isLegalEntity: boolean;
    organizationType: string;
    // Sub-resources
    contactMediums: Partial<ContactMedium>[];
    identifications: Partial<Identification>[];
    taxExemptions: Partial<TaxExemption>[];
    externalReferences: Partial<ExternalReference>[];
    attachments: Partial<Attachment>[];
    relatedParties: Partial<RelatedParty>[];
}

const initialState: FormState = {
    type: 'Individual',
    givenName: '',
    familyName: '',
    middleName: '',
    birthDate: '',
    gender: '',
    tradingName: '',
    isLegalEntity: true,
    organizationType: '',
    contactMediums: [],
    identifications: [],
    taxExemptions: [],
    externalReferences: [],
    attachments: [],
    relatedParties: [],
};

export default function PartyFormPage() {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const isEdit = Boolean(id);

    const { data: existingParty, isLoading: loadingParty } = useParty(id);
    const createMutation = useCreateParty();
    const updateMutation = useUpdateParty();

    const [form, setForm] = useState<FormState>(() => {
        if (existingParty) {
            return {
                type: existingParty['@type'],
                givenName: existingParty['@type'] === 'Individual' ? existingParty.givenName : '',
                familyName: existingParty['@type'] === 'Individual' ? existingParty.familyName : '',
                middleName: existingParty['@type'] === 'Individual' ? existingParty.middleName || '' : '',
                birthDate: existingParty['@type'] === 'Individual' ? existingParty.birthDate || '' : '',
                gender: existingParty['@type'] === 'Individual' ? existingParty.gender || '' : '',
                tradingName: existingParty['@type'] === 'Organization' ? existingParty.tradingName : '',
                isLegalEntity: existingParty['@type'] === 'Organization' ? existingParty.isLegalEntity : true,
                organizationType: existingParty['@type'] === 'Organization' ? existingParty.organizationType || '' : '',
                contactMediums: existingParty.contactMediums || [],
                identifications: existingParty.identifications || [],
                taxExemptions: existingParty.taxExemptions || [],
                externalReferences: existingParty.externalReferences || [],
                attachments: existingParty.attachments || [],
                relatedParties: existingParty.relatedParties || [],
            };
        }
        return initialState;
    });

    useEffect(() => {
        if (existingParty) {
            // eslint-disable-next-line react-hooks/set-state-in-effect
            setForm({
                type: existingParty['@type'],
                givenName: existingParty['@type'] === 'Individual' ? existingParty.givenName : '',
                familyName: existingParty['@type'] === 'Individual' ? existingParty.familyName : '',
                middleName: existingParty['@type'] === 'Individual' ? existingParty.middleName || '' : '',
                birthDate: existingParty['@type'] === 'Individual' ? existingParty.birthDate || '' : '',
                gender: existingParty['@type'] === 'Individual' ? existingParty.gender || '' : '',
                tradingName: existingParty['@type'] === 'Organization' ? existingParty.tradingName : '',
                isLegalEntity: existingParty['@type'] === 'Organization' ? existingParty.isLegalEntity : true,
                organizationType: existingParty['@type'] === 'Organization' ? existingParty.organizationType || '' : '',
                contactMediums: existingParty.contactMediums || [],
                identifications: existingParty.identifications || [],
                taxExemptions: existingParty.taxExemptions || [],
                externalReferences: existingParty.externalReferences || [],
                attachments: existingParty.attachments || [],
                relatedParties: existingParty.relatedParties || [],
            });
        }
    }, [existingParty]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        try {
            if (isEdit && id) {
                const payload: UpdatePartyPayload = form.type === 'Individual'
                    ? {
                        id,
                        '@type': 'Individual',
                        givenName: form.givenName,
                        familyName: form.familyName,
                        middleName: form.middleName || undefined,
                        birthDate: form.birthDate || undefined,
                        gender: form.gender || undefined,
                        contactMediums: form.contactMediums.length > 0 ? form.contactMediums as Omit<ContactMedium, 'id'>[] : undefined,
                        identifications: form.identifications.length > 0 ? form.identifications as Omit<Identification, 'id'>[] : undefined,
                        taxExemptions: form.taxExemptions.length > 0 ? form.taxExemptions as Omit<TaxExemption, 'id'>[] : undefined,
                        externalReferences: form.externalReferences.length > 0 ? form.externalReferences as Omit<ExternalReference, 'id' | 'partyId'>[] : undefined,
                        attachments: form.attachments.length > 0 ? form.attachments as Omit<Attachment, 'id' | 'ownerId'>[] : undefined,
                        relatedParties: form.relatedParties.length > 0 ? form.relatedParties as Omit<RelatedParty, 'id'>[] : undefined,
                    }
                    : {
                        id,
                        '@type': 'Organization',
                        tradingName: form.tradingName,
                        isLegalEntity: form.isLegalEntity,
                        organizationType: form.organizationType || undefined,
                        contactMediums: form.contactMediums.length > 0 ? form.contactMediums as Omit<ContactMedium, 'id'>[] : undefined,
                        identifications: form.identifications.length > 0 ? form.identifications as Omit<Identification, 'id'>[] : undefined,
                        taxExemptions: form.taxExemptions.length > 0 ? form.taxExemptions as Omit<TaxExemption, 'id'>[] : undefined,
                        externalReferences: form.externalReferences.length > 0 ? form.externalReferences as Omit<ExternalReference, 'id' | 'partyId'>[] : undefined,
                        attachments: form.attachments.length > 0 ? form.attachments as Omit<Attachment, 'id' | 'ownerId'>[] : undefined,
                        relatedParties: form.relatedParties.length > 0 ? form.relatedParties as Omit<RelatedParty, 'id'>[] : undefined,
                    };
                await updateMutation.mutateAsync(payload);
            } else {
                const payload: CreatePartyPayload = form.type === 'Individual'
                    ? {
                        '@type': 'Individual',
                        givenName: form.givenName,
                        familyName: form.familyName,
                        middleName: form.middleName || undefined,
                        birthDate: form.birthDate || undefined,
                        gender: form.gender || undefined,
                        contactMediums: form.contactMediums.length > 0 ? form.contactMediums as Omit<ContactMedium, 'id'>[] : undefined,
                        identifications: form.identifications.length > 0 ? form.identifications as Omit<Identification, 'id'>[] : undefined,
                        taxExemptions: form.taxExemptions.length > 0 ? form.taxExemptions as Omit<TaxExemption, 'id'>[] : undefined,
                        externalReferences: form.externalReferences.length > 0 ? form.externalReferences as Omit<ExternalReference, 'id' | 'partyId'>[] : undefined,
                        attachments: form.attachments.length > 0 ? form.attachments as Omit<Attachment, 'id' | 'ownerId'>[] : undefined,
                        relatedParties: form.relatedParties.length > 0 ? form.relatedParties as Omit<RelatedParty, 'id'>[] : undefined,
                    }
                    : {
                        '@type': 'Organization',
                        tradingName: form.tradingName,
                        isLegalEntity: form.isLegalEntity,
                        organizationType: form.organizationType || undefined,
                        contactMediums: form.contactMediums.length > 0 ? form.contactMediums as Omit<ContactMedium, 'id'>[] : undefined,
                        identifications: form.identifications.length > 0 ? form.identifications as Omit<Identification, 'id'>[] : undefined,
                        taxExemptions: form.taxExemptions.length > 0 ? form.taxExemptions as Omit<TaxExemption, 'id'>[] : undefined,
                        externalReferences: form.externalReferences.length > 0 ? form.externalReferences as Omit<ExternalReference, 'id' | 'partyId'>[] : undefined,
                        attachments: form.attachments.length > 0 ? form.attachments as Omit<Attachment, 'id' | 'ownerId'>[] : undefined,
                        relatedParties: form.relatedParties.length > 0 ? form.relatedParties as Omit<RelatedParty, 'id'>[] : undefined,
                    };
                await createMutation.mutateAsync(payload);
            }
            navigate('/parties');
        } catch (err) {
            console.error('Failed to save party:', err);
        }
    };

    const addContactMedium = () => {
        setForm((prev) => ({
            ...prev,
            contactMediums: [...prev.contactMediums, { mediumType: 'email', preferred: false, value: '' }],
        }));
    };

    const removeContactMedium = (index: number) => {
        setForm((prev) => ({
            ...prev,
            contactMediums: prev.contactMediums.filter((_, i) => i !== index),
        }));
    };

    const addIdentification = () => {
        setForm((prev) => ({
            ...prev,
            identifications: [...prev.identifications, { identificationType: '', identificationId: '' }],
        }));
    };

    const removeIdentification = (index: number) => {
        setForm((prev) => ({
            ...prev,
            identifications: prev.identifications.filter((_, i) => i !== index),
        }));
    };

    const addTaxExemption = () => {
        setForm((prev) => ({
            ...prev,
            taxExemptions: [...prev.taxExemptions, {
                certificateNumber: '',
                issuingJurisdiction: '',
                validFor: { startDateTime: new Date().toISOString() }
            }],
        }));
    };

    const removeTaxExemption = (index: number) => {
        setForm((prev) => ({
            ...prev,
            taxExemptions: prev.taxExemptions.filter((_, i) => i !== index),
        }));
    };

    const addExternalReference = () => {
        setForm((prev) => ({
            ...prev,
            externalReferences: [...prev.externalReferences, { externalSystemId: '', externalReference: '' }],
        }));
    };

    const removeExternalReference = (index: number) => {
        setForm((prev) => ({
            ...prev,
            externalReferences: prev.externalReferences.filter((_, i) => i !== index),
        }));
    };

    const addAttachment = () => {
        setForm((prev) => ({
            ...prev,
            attachments: [...prev.attachments, { name: '', mimeType: '', url: '', type: 'Document' }],
        }));
    };

    const removeAttachment = (index: number) => {
        setForm((prev) => ({
            ...prev,
            attachments: prev.attachments.filter((_, i) => i !== index),
        }));
    };

    const addRelatedParty = () => {
        setForm((prev) => ({
            ...prev,
            relatedParties: [...prev.relatedParties, { role: '', relatedPartyId: '', permissions: [] }],
        }));
    };

    const removeRelatedParty = (index: number) => {
        setForm((prev) => ({
            ...prev,
            relatedParties: prev.relatedParties.filter((_, i) => i !== index),
        }));
    };

    const isPending = createMutation.isPending || updateMutation.isPending;

    if (isEdit && loadingParty) {
        return (
            <div className="party-form-page">
                <div className="loading-state">
                    <Loader2 className="spin" size={32} />
                    <p>Loading party...</p>
                </div>
            </div>
        );
    }

    return (
        <div className="party-form-page">
            <div className="page-header">
                <div className="page-header-content">
                    <Link to="/parties" className="back-link">
                        <ArrowLeft size={18} />
                        <span>Back to Parties</span>
                    </Link>
                    <h2>{isEdit ? 'Edit Party' : 'Create Party'}</h2>
                    <p className="page-description">
                        {isEdit ? `Editing party ${id}` : 'Register a new Individual or Organization'}
                    </p>
                </div>
                <button
                    type="submit"
                    form="party-form"
                    className="btn btn-primary"
                    disabled={isPending}
                >
                    {isPending ? <Loader2 className="spin" size={18} /> : <Save size={18} />}
                    <span>{isEdit ? 'Save Changes' : 'Create Party'}</span>
                </button>
            </div>

            <form id="party-form" className="form-container" onSubmit={handleSubmit}>
                {/* Party Type Selector */}
                <div className="card form-section">
                    <h3>Party Type</h3>
                    <div className="type-selector">
                        <label className={`type-option ${form.type === 'Individual' ? 'active' : ''}`}>
                            <input
                                type="radio"
                                name="type"
                                value="Individual"
                                checked={form.type === 'Individual'}
                                onChange={() => setForm((prev) => ({ ...prev, type: 'Individual' }))}
                            />
                            <span className="type-label">Individual</span>
                            <span className="type-desc">A natural person</span>
                        </label>
                        <label className={`type-option ${form.type === 'Organization' ? 'active' : ''}`}>
                            <input
                                type="radio"
                                name="type"
                                value="Organization"
                                checked={form.type === 'Organization'}
                                onChange={() => setForm((prev) => ({ ...prev, type: 'Organization' }))}
                            />
                            <span className="type-label">Organization</span>
                            <span className="type-desc">A company or group</span>
                        </label>
                    </div>
                </div>

                {/* Basic Info */}
                <div className="card form-section">
                    <h3>{form.type === 'Individual' ? 'Personal Information' : 'Organization Information'}</h3>

                    {form.type === 'Individual' ? (
                        <div className="form-grid">
                            <div className="form-group">
                                <label htmlFor="givenName">Given Name *</label>
                                <input
                                    id="givenName"
                                    type="text"
                                    value={form.givenName}
                                    onChange={(e) => setForm((prev) => ({ ...prev, givenName: e.target.value }))}
                                    required
                                    placeholder="John"
                                />
                            </div>
                            <div className="form-group">
                                <label htmlFor="familyName">Family Name *</label>
                                <input
                                    id="familyName"
                                    type="text"
                                    value={form.familyName}
                                    onChange={(e) => setForm((prev) => ({ ...prev, familyName: e.target.value }))}
                                    required
                                    placeholder="Doe"
                                />
                            </div>
                            <div className="form-group">
                                <label htmlFor="middleName">Middle Name</label>
                                <input
                                    id="middleName"
                                    type="text"
                                    value={form.middleName}
                                    onChange={(e) => setForm((prev) => ({ ...prev, middleName: e.target.value }))}
                                    placeholder="Optional"
                                />
                            </div>
                            <div className="form-group">
                                <label htmlFor="birthDate">Birth Date</label>
                                <input
                                    id="birthDate"
                                    type="date"
                                    value={form.birthDate}
                                    onChange={(e) => setForm((prev) => ({ ...prev, birthDate: e.target.value }))}
                                />
                            </div>
                            <div className="form-group">
                                <label htmlFor="gender">Gender</label>
                                <select
                                    id="gender"
                                    value={form.gender}
                                    onChange={(e) => setForm((prev) => ({ ...prev, gender: e.target.value }))}
                                >
                                    <option value="">Select...</option>
                                    <option value="male">Male</option>
                                    <option value="female">Female</option>
                                    <option value="other">Other</option>
                                </select>
                            </div>
                        </div>
                    ) : (
                        <div className="form-grid">
                            <div className="form-group form-group--full">
                                <label htmlFor="tradingName">Trading Name *</label>
                                <input
                                    id="tradingName"
                                    type="text"
                                    value={form.tradingName}
                                    onChange={(e) => setForm((prev) => ({ ...prev, tradingName: e.target.value }))}
                                    required
                                    placeholder="Acme Corporation"
                                />
                            </div>
                            <div className="form-group">
                                <label htmlFor="organizationType">Organization Type</label>
                                <input
                                    id="organizationType"
                                    type="text"
                                    value={form.organizationType}
                                    onChange={(e) => setForm((prev) => ({ ...prev, organizationType: e.target.value }))}
                                    placeholder="e.g., LLC, Inc, GmbH"
                                />
                            </div>
                            <div className="form-group">
                                <label className="checkbox-label">
                                    <input
                                        type="checkbox"
                                        checked={form.isLegalEntity}
                                        onChange={(e) => setForm((prev) => ({ ...prev, isLegalEntity: e.target.checked }))}
                                    />
                                    <span>Legal Entity</span>
                                </label>
                            </div>
                        </div>
                    )}
                </div>

                {/* Contact Mediums */}
                <div className="card form-section">
                    <div className="section-header">
                        <h3>Contact Mediums</h3>
                        <button type="button" className="btn btn-secondary btn-sm" onClick={addContactMedium}>
                            <Plus size={16} />
                            <span>Add Contact</span>
                        </button>
                    </div>

                    {form.contactMediums.length === 0 ? (
                        <p className="empty-text">No contact mediums added</p>
                    ) : (
                        <div className="repeatable-list">
                            {form.contactMediums.map((contact, index) => (
                                <div key={index} className="repeatable-item">
                                    <div className="form-grid">
                                        <div className="form-group">
                                            <label>Type</label>
                                            <select
                                                value={contact.mediumType || 'email'}
                                                onChange={(e) => {
                                                    const newContacts = [...form.contactMediums];
                                                    newContacts[index] = { ...newContacts[index], mediumType: e.target.value as 'email' | 'phone' | 'postal' };
                                                    setForm((prev) => ({ ...prev, contactMediums: newContacts }));
                                                }}
                                            >
                                                <option value="email">Email</option>
                                                <option value="phone">Phone</option>
                                                <option value="postal">Postal</option>
                                            </select>
                                        </div>
                                        <div className="form-group">
                                            <label>Value</label>
                                            <input
                                                type="text"
                                                value={contact.value || ''}
                                                onChange={(e) => {
                                                    const newContacts = [...form.contactMediums];
                                                    newContacts[index] = { ...newContacts[index], value: e.target.value };
                                                    setForm((prev) => ({ ...prev, contactMediums: newContacts }));
                                                }}
                                                placeholder={contact.mediumType === 'email' ? 'email@example.com' : '+1234567890'}
                                            />
                                        </div>
                                        <div className="form-group form-group--action">
                                            <button
                                                type="button"
                                                className="btn-icon btn-icon--danger"
                                                onClick={() => removeContactMedium(index)}
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

                {/* Identifications */}
                <div className="card form-section">
                    <div className="section-header">
                        <h3>Identifications</h3>
                        <button type="button" className="btn btn-secondary btn-sm" onClick={addIdentification}>
                            <Plus size={16} />
                            <span>Add ID</span>
                        </button>
                    </div>

                    {form.identifications.length === 0 ? (
                        <p className="empty-text">No identifications added</p>
                    ) : (
                        <div className="repeatable-list">
                            {form.identifications.map((ident, index) => (
                                <div key={index} className="repeatable-item">
                                    <div className="form-grid">
                                        <div className="form-group">
                                            <label>ID Type</label>
                                            <input
                                                type="text"
                                                value={ident.identificationType || ''}
                                                onChange={(e) => {
                                                    const newIdents = [...form.identifications];
                                                    newIdents[index] = { ...newIdents[index], identificationType: e.target.value };
                                                    setForm((prev) => ({ ...prev, identifications: newIdents }));
                                                }}
                                                placeholder="e.g., Passport, SSN"
                                            />
                                        </div>
                                        <div className="form-group">
                                            <label>ID Number</label>
                                            <input
                                                type="text"
                                                value={ident.identificationId || ''}
                                                onChange={(e) => {
                                                    const newIdents = [...form.identifications];
                                                    newIdents[index] = { ...newIdents[index], identificationId: e.target.value };
                                                    setForm((prev) => ({ ...prev, identifications: newIdents }));
                                                }}
                                                placeholder="ID number"
                                            />
                                        </div>
                                        <div className="form-group form-group--action">
                                            <button
                                                type="button"
                                                className="btn-icon btn-icon--danger"
                                                onClick={() => removeIdentification(index)}
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

                {/* Tax Exemptions */}
                <div className="card form-section">
                    <div className="section-header">
                        <h3>Tax Exemptions</h3>
                        <button type="button" className="btn btn-secondary btn-sm" onClick={addTaxExemption}>
                            <Plus size={16} />
                            <span>Add Exemption</span>
                        </button>
                    </div>

                    {form.taxExemptions.length === 0 ? (
                        <p className="empty-text">No tax exemptions added</p>
                    ) : (
                        <div className="repeatable-list">
                            {form.taxExemptions.map((ex, index) => (
                                <div key={index} className="repeatable-item">
                                    <div className="form-grid">
                                        <div className="form-group">
                                            <label>Certificate Number</label>
                                            <input
                                                type="text"
                                                value={ex.certificateNumber || ''}
                                                onChange={(e) => {
                                                    const newExemptions = [...form.taxExemptions];
                                                    newExemptions[index] = { ...newExemptions[index], certificateNumber: e.target.value };
                                                    setForm((prev) => ({ ...prev, taxExemptions: newExemptions }));
                                                }}
                                                placeholder="Certificate #"
                                            />
                                        </div>
                                        <div className="form-group">
                                            <label>Jurisdiction</label>
                                            <input
                                                type="text"
                                                value={ex.issuingJurisdiction || ''}
                                                onChange={(e) => {
                                                    const newExemptions = [...form.taxExemptions];
                                                    newExemptions[index] = { ...newExemptions[index], issuingJurisdiction: e.target.value };
                                                    setForm((prev) => ({ ...prev, taxExemptions: newExemptions }));
                                                }}
                                                placeholder="Issuing Jurisdiction"
                                            />
                                        </div>
                                        <div className="form-group">
                                            <label>Start Date</label>
                                            <input
                                                type="date"
                                                value={ex.validFor?.startDateTime ? new Date(ex.validFor.startDateTime).toISOString().split('T')[0] : ''}
                                                onChange={(e) => {
                                                    const newExemptions = [...form.taxExemptions];
                                                    const currentValidFor = newExemptions[index].validFor || {};
                                                    newExemptions[index] = {
                                                        ...newExemptions[index],
                                                        validFor: { ...currentValidFor, startDateTime: new Date(e.target.value).toISOString() }
                                                    };
                                                    setForm((prev) => ({ ...prev, taxExemptions: newExemptions }));
                                                }}
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

                {/* External References */}
                <div className="card form-section">
                    <div className="section-header">
                        <h3>External References</h3>
                        <button type="button" className="btn btn-secondary btn-sm" onClick={addExternalReference}>
                            <Plus size={16} />
                            <span>Add Reference</span>
                        </button>
                    </div>

                    {form.externalReferences.length === 0 ? (
                        <p className="empty-text">No external references added</p>
                    ) : (
                        <div className="repeatable-list">
                            {form.externalReferences.map((ref, index) => (
                                <div key={index} className="repeatable-item">
                                    <div className="form-grid">
                                        <div className="form-group">
                                            <label>System ID</label>
                                            <input
                                                type="text"
                                                value={ref.externalSystemId || ''}
                                                onChange={(e) => {
                                                    const newRefs = [...form.externalReferences];
                                                    newRefs[index] = { ...newRefs[index], externalSystemId: e.target.value };
                                                    setForm((prev) => ({ ...prev, externalReferences: newRefs }));
                                                }}
                                                placeholder="e.g. LegacyCRM"
                                            />
                                        </div>
                                        <div className="form-group">
                                            <label>Reference ID</label>
                                            <input
                                                type="text"
                                                value={ref.externalReference || ''}
                                                onChange={(e) => {
                                                    const newRefs = [...form.externalReferences];
                                                    newRefs[index] = { ...newRefs[index], externalReference: e.target.value };
                                                    setForm((prev) => ({ ...prev, externalReferences: newRefs }));
                                                }}
                                                placeholder="e.g. CUST-123"
                                            />
                                        </div>
                                        <div className="form-group form-group--action">
                                            <button
                                                type="button"
                                                className="btn-icon btn-icon--danger"
                                                onClick={() => removeExternalReference(index)}
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

                {/* Attachments */}
                <div className="card form-section">
                    <div className="section-header">
                        <h3>Attachments</h3>
                        <button type="button" className="btn btn-secondary btn-sm" onClick={addAttachment}>
                            <Plus size={16} />
                            <span>Add Attachment</span>
                        </button>
                    </div>

                    {form.attachments.length === 0 ? (
                        <p className="empty-text">No attachments added</p>
                    ) : (
                        <div className="repeatable-list">
                            {form.attachments.map((att, index) => (
                                <div key={index} className="repeatable-item">
                                    <div className="form-grid">
                                        <div className="form-group">
                                            <label>Name</label>
                                            <input
                                                type="text"
                                                value={att.name || ''}
                                                onChange={(e) => {
                                                    const newAtts = [...form.attachments];
                                                    newAtts[index] = { ...newAtts[index], name: e.target.value };
                                                    setForm((prev) => ({ ...prev, attachments: newAtts }));
                                                }}
                                                placeholder="Document Name"
                                            />
                                        </div>
                                        <div className="form-group">
                                            <label>URL</label>
                                            <input
                                                type="text"
                                                value={att.url || ''}
                                                onChange={(e) => {
                                                    const newAtts = [...form.attachments];
                                                    newAtts[index] = { ...newAtts[index], url: e.target.value };
                                                    setForm((prev) => ({ ...prev, attachments: newAtts }));
                                                }}
                                                placeholder="https://..."
                                            />
                                        </div>
                                        <div className="form-group">
                                            <label>Mime Type</label>
                                            <input
                                                type="text"
                                                value={att.mimeType || ''}
                                                onChange={(e) => {
                                                    const newAtts = [...form.attachments];
                                                    newAtts[index] = { ...newAtts[index], mimeType: e.target.value };
                                                    setForm((prev) => ({ ...prev, attachments: newAtts }));
                                                }}
                                                placeholder="application/pdf"
                                            />
                                        </div>
                                        <div className="form-group form-group--action">
                                            <button
                                                type="button"
                                                className="btn-icon btn-icon--danger"
                                                onClick={() => removeAttachment(index)}
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

                {/* Related Parties */}
                <div className="card form-section">
                    <div className="section-header">
                        <h3>Related Parties</h3>
                        <button type="button" className="btn btn-secondary btn-sm" onClick={addRelatedParty}>
                            <Plus size={16} />
                            <span>Add Relationship</span>
                        </button>
                    </div>

                    {form.relatedParties.length === 0 ? (
                        <p className="empty-text">No related parties added</p>
                    ) : (
                        <div className="repeatable-list">
                            {form.relatedParties.map((rel, index) => (
                                <div key={index} className="repeatable-item">
                                    <div className="form-grid">
                                        <div className="form-group">
                                            <label>Party</label>
                                            <PartySelector
                                                onSelect={(party) => {
                                                    const newRels = [...form.relatedParties];
                                                    newRels[index] = {
                                                        ...newRels[index],
                                                        relatedPartyId: party.id,
                                                        // Optionally set name here if backend preserved it, but ID is critical
                                                    };
                                                    setForm((prev) => ({ ...prev, relatedParties: newRels }));
                                                }}
                                            />
                                            {rel.relatedPartyId && <div className="selected-value">Selected ID: {rel.relatedPartyId}</div>}
                                        </div>
                                        <div className="form-group">
                                            <label>Role</label>
                                            <input
                                                type="text"
                                                value={rel.role || ''}
                                                onChange={(e) => {
                                                    const newRels = [...form.relatedParties];
                                                    newRels[index] = { ...newRels[index], role: e.target.value };
                                                    setForm((prev) => ({ ...prev, relatedParties: newRels }));
                                                }}
                                                placeholder="e.g. Employee, Next of Kin"
                                            />
                                        </div>
                                        <div className="form-group">
                                            <label>Permissions (comma separated)</label>
                                            <input
                                                type="text"
                                                value={rel.permissions?.join(', ') || ''}
                                                onChange={(e) => {
                                                    const newRels = [...form.relatedParties];
                                                    const perms = e.target.value.split(',').map(s => s.trim()).filter(Boolean);
                                                    newRels[index] = { ...newRels[index], permissions: perms };
                                                    setForm((prev) => ({ ...prev, relatedParties: newRels }));
                                                }}
                                                placeholder="read, write"
                                            />
                                        </div>
                                        <div className="form-group form-group--action">
                                            <button
                                                type="button"
                                                className="btn-icon btn-icon--danger"
                                                onClick={() => removeRelatedParty(index)}
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
