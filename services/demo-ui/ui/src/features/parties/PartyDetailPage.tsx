import { useState, useCallback } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { ArrowLeft, Edit, Trash2, Mail, Phone, MapPin, CreditCard, Users, Loader2 } from 'lucide-react';
import { useParty, useDeleteParty } from './api';
import { getPartyDisplayName, isIndividual } from './types';
import { useNotification } from '../../components/common/Toast';
import './PartyDetailPage.css';

export default function PartyDetailPage() {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const { data: party, isLoading, error } = useParty(id);
    const deleteMutation = useDeleteParty();
    const { showToast } = useNotification();
    const [isDeleting, setIsDeleting] = useState(false);

    const checkDeletionStatus = useCallback(async (partyId: string, attempts = 0) => {
        if (attempts > 15) {
            setIsDeleting(false);
            showToast('Deletion operation timed out. Please check status manually.', 'info');
            navigate('/parties');
            return;
        }

        // Add initial delay to let the saga start processing
        if (attempts === 0) {
            await new Promise(resolve => setTimeout(resolve, 500));
        }

        try {
            const { apiClient } = await import('../../api/client');
            const response = await apiClient.get(`/api/parties/${partyId}`);
            const partyData = response.data;

            if (partyData.status === 'Deleted') {
                setIsDeleting(false);
                showToast('Party deleted successfully', 'success');
                navigate('/parties');
            } else if (partyData.status === 'Active') {
                setIsDeleting(false);
                showToast('Deletion failed: Party has active linked customers.', 'error');
                // Stay on page - don't navigate
            } else if (partyData.status === 'DeletionPending') {
                setTimeout(() => checkDeletionStatus(partyId, attempts + 1), 1000);
            } else {
                setIsDeleting(false);
                showToast(`Deletion ended with status: ${partyData.status}`, 'info');
                navigate('/parties');
            }
        } catch (err: unknown) {
            // 404 means party was deleted
            if (err && typeof err === 'object' && 'response' in err) {
                const axiosErr = err as { response?: { status?: number } };
                if (axiosErr.response?.status === 404) {
                    setIsDeleting(false);
                    showToast('Party deleted successfully', 'success');
                    navigate('/parties');
                    return;
                }
            }
            console.error("Error checking status", err);
            setIsDeleting(false);
            showToast('Error checking deletion status', 'error');
        }
    }, [showToast, navigate]);

    const handleDelete = () => {
        if (party && confirm(`Are you sure you want to delete "${getPartyDisplayName(party)}"?`)) {
            setIsDeleting(true);
            deleteMutation.mutate(party.id, {
                onSuccess: () => {
                    showToast('Deletion initiated...', 'info');
                    checkDeletionStatus(party.id);
                },
                onError: (err) => {
                    setIsDeleting(false);
                    showToast(`Failed to delete: ${err.message}`, 'error');
                },
            });
        }
    };

    if (isLoading) {
        return (
            <div className="party-detail-page">
                <div className="loading-state" role="status">
                    <Loader2 className="spin" size={32} />
                    <p>Loading party details...</p>
                </div>
            </div>
        );
    }

    if (error || !party) {
        return (
            <div className="party-detail-page">
                <div className="error-state card" role="alert">
                    <p>Failed to load party: {error?.message || 'Party not found'}</p>
                    <Link to="/parties" className="btn btn-secondary">
                        Back to Parties
                    </Link>
                </div>
            </div>
        );
    }

    return (
        <div className="party-detail-page">
            <div className="page-header">
                <div className="page-header-content">
                    <Link to="/parties" className="back-link">
                        <ArrowLeft size={18} />
                        <span>Back to Parties</span>
                    </Link>
                    <h2>{getPartyDisplayName(party)}</h2>
                    <div className="party-meta">
                        <span className={`party-type-badge ${party['@type'].toLowerCase()}`}>
                            {party['@type']}
                        </span>
                        <span className={`status-badge ${party.status}`}>
                            {party.status}
                        </span>
                    </div>
                </div>
                <div className="page-actions">
                    <Link to={`/parties/${id}/edit`} className="btn btn-secondary">
                        <Edit size={18} />
                        <span>Edit</span>
                    </Link>
                    <button
                        className="btn btn-danger"
                        onClick={handleDelete}
                        disabled={deleteMutation.isPending || isDeleting}
                    >
                        <Trash2 size={18} />
                        <span>Delete</span>
                    </button>
                </div>
            </div>

            <div className="detail-grid">
                {/* Basic Info Card */}
                <div className="card detail-card">
                    <h3>Basic Information</h3>
                    <dl className="detail-list">
                        <div className="detail-item">
                            <dt>ID</dt>
                            <dd className="mono">{party.id}</dd>
                        </div>
                        <div className="detail-item">
                            <dt>Type</dt>
                            <dd>{party['@type']}</dd>
                        </div>
                        {isIndividual(party) ? (
                            <>
                                <div className="detail-item">
                                    <dt>Given Name</dt>
                                    <dd>{party.givenName}</dd>
                                </div>
                                <div className="detail-item">
                                    <dt>Family Name</dt>
                                    <dd>{party.familyName}</dd>
                                </div>
                                {party.birthDate && (
                                    <div className="detail-item">
                                        <dt>Birth Date</dt>
                                        <dd>{party.birthDate}</dd>
                                    </div>
                                )}
                                {party.gender && (
                                    <div className="detail-item">
                                        <dt>Gender</dt>
                                        <dd>{party.gender}</dd>
                                    </div>
                                )}
                            </>
                        ) : (
                            <>
                                <div className="detail-item">
                                    <dt>Trading Name</dt>
                                    <dd>{party.tradingName}</dd>
                                </div>
                                <div className="detail-item">
                                    <dt>Legal Entity</dt>
                                    <dd>{party.isLegalEntity ? 'Yes' : 'No'}</dd>
                                </div>
                                {party.organizationType && (
                                    <div className="detail-item">
                                        <dt>Organization Type</dt>
                                        <dd>{party.organizationType}</dd>
                                    </div>
                                )}
                            </>
                        )}
                    </dl>
                </div>

                {/* Contact Mediums Card */}
                <div className="card detail-card">
                    <h3>
                        <span>Contact Mediums</span>
                        <span className="count-badge">{party.contactMediums?.length || 0}</span>
                    </h3>
                    {party.contactMediums && party.contactMediums.length > 0 ? (
                        <ul className="contact-list">
                            {party.contactMediums.map((contact) => (
                                <li key={contact.id} className="contact-item">
                                    <div className="contact-icon">
                                        {contact.mediumType === 'email' && <Mail size={18} />}
                                        {contact.mediumType === 'phone' && <Phone size={18} />}
                                        {contact.mediumType === 'postal' && <MapPin size={18} />}
                                    </div>
                                    <div className="contact-content">
                                        <span className="contact-type">{contact.mediumType}</span>
                                        {contact.value && <span className="contact-value">{contact.value}</span>}
                                        {contact.mediumType === 'postal' && (
                                            <span className="contact-address">
                                                {[contact.street1, contact.city, contact.country].filter(Boolean).join(', ')}
                                            </span>
                                        )}
                                    </div>
                                    {contact.preferred && <span className="preferred-badge">Preferred</span>}
                                </li>
                            ))}
                        </ul>
                    ) : (
                        <p className="empty-text">No contact mediums</p>
                    )}
                </div>

                {/* Identifications Card */}
                <div className="card detail-card">
                    <h3>
                        <span>Identifications</span>
                        <span className="count-badge">{party.identifications?.length || 0}</span>
                    </h3>
                    {party.identifications && party.identifications.length > 0 ? (
                        <ul className="id-list">
                            {party.identifications.map((ident) => (
                                <li key={ident.id} className="id-item">
                                    <CreditCard size={18} className="id-icon" />
                                    <div className="id-content">
                                        <span className="id-type">{ident.identificationType}</span>
                                        <span className="id-value">{ident.identificationId}</span>
                                        {ident.issuingAuthority && (
                                            <span className="id-issuer">Issued by: {ident.issuingAuthority}</span>
                                        )}
                                    </div>
                                </li>
                            ))}
                        </ul>
                    ) : (
                        <p className="empty-text">No identifications</p>
                    )}
                </div>

                {/* Related Parties Card */}
                <div className="card detail-card">
                    <h3>
                        <span>Related Parties</span>
                        <span className="count-badge">{party.relatedParties?.length || 0}</span>
                    </h3>
                    {party.relatedParties && party.relatedParties.length > 0 ? (
                        <ul className="related-list">
                            {party.relatedParties.map((related) => (
                                <li key={related.id} className="related-item">
                                    <Users size={18} className="related-icon" />
                                    <div className="related-content">
                                        <span className="related-name">{related.relatedPartyName || related.relatedPartyId}</span>
                                        <span className="related-role">{related.role}</span>
                                    </div>
                                    <Link to={`/parties/${related.relatedPartyId}`} className="btn btn-link">
                                        View
                                    </Link>
                                </li>
                            ))}
                        </ul>
                    ) : (
                        <p className="empty-text">No related parties</p>
                    )}
                </div>
            </div>
        </div>
    );
}
