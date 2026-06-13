import { useParams, Link } from 'react-router-dom';
import { ArrowLeft, Edit, Loader2, Tag, Clock, CheckCircle, XCircle, AlertCircle, DollarSign, Paperclip, Layers } from 'lucide-react';
import { useOffering } from './api';
import type { ProductOfferingPrice } from './types';
import './SpecificationListPage.css';

const PRICE_TYPE_LABELS: Record<ProductOfferingPrice['priceType'], string> = {
    recurring: 'Recurring',
    one_time: 'One-time',
    usage: 'Usage',
};

function formatMoney(price: ProductOfferingPrice): string {
    const { unit, value } = price.price;
    const amount = new Intl.NumberFormat(undefined, {
        style: 'currency',
        currency: unit || 'USD',
    }).format(value);
    return price.unitOfMeasure ? `${amount} / ${price.unitOfMeasure}` : amount;
}

export default function OfferingDetailPage() {
    const { id } = useParams<{ id: string }>();
    const { data: offering, isLoading, error } = useOffering(id);

    if (isLoading) {
        return (
            <div className="specification-detail-page">
                <div className="loading-state">
                    <Loader2 className="spin" size={32} />
                    <p>Loading offering...</p>
                </div>
            </div>
        );
    }

    if (error || !offering) {
        return (
            <div className="specification-detail-page">
                <div className="error-state">
                    <AlertCircle size={48} className="error-icon" />
                    <p>{error ? `Error: ${error.message}` : 'Offering not found'}</p>
                    <Link to="/catalog/offerings" className="btn btn-primary">
                        Back to Offerings
                    </Link>
                </div>
            </div>
        );
    }

    const getStatusIcon = (status: string) => {
        switch (status) {
            case 'Active': return <CheckCircle size={16} className="text-success" />;
            case 'Retired': return <XCircle size={16} className="text-danger" />;
            case 'Draft': return <Clock size={16} className="text-muted" />;
            default: return <AlertCircle size={16} />;
        }
    };

    const prices = offering.productOfferingPrice || [];
    const categories = offering.categories || [];
    const attachments = offering.attachments || [];

    return (
        <div className="specification-detail-page">
            <div className="page-header">
                <div className="page-header-content">
                    <Link to="/catalog/offerings" className="back-link">
                        <ArrowLeft size={18} />
                        <span>Back to Offerings</span>
                    </Link>
                    <div className="header-title-area mt-2">
                        <div className="title-with-badge">
                            <h2>{offering.name}</h2>
                            {offering.isBundle && <span className="bundle-badge">Bundle</span>}
                            {offering.isSellable && <span className="badge badge-success-outline">Sellable</span>}
                        </div>
                        {offering.productSpecification && (
                            <p className="lead text-muted">{offering.productSpecification.name}</p>
                        )}
                    </div>
                </div>
                <Link to={`/catalog/offerings/${id}/edit`} className="btn btn-secondary">
                    <Edit size={18} />
                    <span>Edit Offering</span>
                </Link>
            </div>

            <div className="detail-grid mt-4">
                <div className="main-content">
                    <div className="card detail-section">
                        <h3>Overview</h3>
                        <div className="overview-content">
                            <p className="description-text">
                                {offering.description || <span className="text-muted">No description provided.</span>}
                            </p>
                            <div className="metadata-grid mt-4">
                                <div className="metadata-item">
                                    <span className="label">Lifecycle Status</span>
                                    <div className="value-with-icon">
                                        {getStatusIcon(offering.lifecycleStatus)}
                                        <span className={`status-badge ${offering.lifecycleStatus.toLowerCase()}`}>
                                            {offering.lifecycleStatus}
                                        </span>
                                    </div>
                                </div>
                                <div className="metadata-item">
                                    <span className="label">Last Updated</span>
                                    <span className="value">{new Date(offering.lastUpdate).toLocaleString()}</span>
                                </div>
                                {offering.validFor && (
                                    <div className="metadata-item">
                                        <span className="label">Validity Period</span>
                                        <span className="value">
                                            {offering.validFor.startDateTime ? new Date(offering.validFor.startDateTime).toLocaleDateString() : 'Start'}
                                            {' - '}
                                            {offering.validFor.endDateTime ? new Date(offering.validFor.endDateTime).toLocaleDateString() : 'End'}
                                        </span>
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>

                    <div className="card detail-section mt-4">
                        <div className="section-header">
                            <h3>Pricing</h3>
                            <span className="badge badge-secondary">{prices.length} Prices</span>
                        </div>
                        {prices.length === 0 ? (
                            <p className="empty-text">No prices defined for this offering.</p>
                        ) : (
                            <div className="characteristics-list mt-3">
                                {prices.map((price, idx) => (
                                    <div key={price.id ?? idx} className="char-item card">
                                        <div className="char-header">
                                            <div className="char-title">
                                                <DollarSign size={16} />
                                                <strong>{formatMoney(price)}</strong>
                                                <span className="char-type-badge">{PRICE_TYPE_LABELS[price.priceType]}</span>
                                            </div>
                                            {price.priceAlteration && (
                                                <span className="badge badge-info-outline">
                                                    {price.priceAlteration.type === 'discount' ? 'Discount' : 'Fee'}: {price.priceAlteration.name}
                                                </span>
                                            )}
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>

                    {categories.length > 0 && (
                        <div className="card detail-section mt-4">
                            <div className="section-header">
                                <h3>Categories</h3>
                                <span className="badge badge-secondary">{categories.length} Items</span>
                            </div>
                            <div className="values-cloud mt-3">
                                {categories.map(cat => (
                                    <span key={cat.id} className="value-tag">
                                        <Layers size={14} /> {cat.name}
                                    </span>
                                ))}
                            </div>
                        </div>
                    )}

                    {attachments.length > 0 && (
                        <div className="card detail-section mt-4">
                            <div className="section-header">
                                <h3>Attachments</h3>
                                <span className="badge badge-secondary">{attachments.length} Items</span>
                            </div>
                            <div className="characteristics-list mt-3">
                                {attachments.map(att => (
                                    <div key={att.id} className="char-item card">
                                        <div className="char-header">
                                            <div className="char-title">
                                                <Paperclip size={16} />
                                                <a href={att.url} target="_blank" rel="noopener noreferrer">{att.name}</a>
                                                <span className="char-type-badge">{att.type}</span>
                                            </div>
                                        </div>
                                        {att.description && (
                                            <p className="char-desc text-muted mt-2">{att.description}</p>
                                        )}
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}
                </div>

                <div className="sidebar-content">
                    <div className="card info-card">
                        <h3>Quick Stats</h3>
                        <div className="stats-list">
                            <div className="stat-item">
                                <span className="stat-label">Sellable</span>
                                <span className="stat-value">{offering.isSellable ? 'Yes' : 'No'}</span>
                            </div>
                            <div className="stat-item">
                                <span className="stat-label">Bundle</span>
                                <span className="stat-value">{offering.isBundle ? 'Yes' : 'No'}</span>
                            </div>
                            <div className="stat-item">
                                <span className="stat-label">Prices</span>
                                <span className="stat-value">{prices.length}</span>
                            </div>
                            <div className="stat-item">
                                <span className="stat-label">Categories</span>
                                <span className="stat-value">{categories.length}</span>
                            </div>
                        </div>
                    </div>

                    {offering.productSpecification && (
                        <div className="card info-card mt-4">
                            <h3>Specification</h3>
                            <div className="stats-list">
                                <div className="stat-item">
                                    <span className="stat-label">
                                        <Tag size={14} /> {offering.productSpecification.name}
                                    </span>
                                </div>
                            </div>
                            <Link
                                to={`/catalog/specifications/${offering.productSpecification.id}`}
                                className="btn btn-secondary mt-3"
                            >
                                View Specification
                            </Link>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}
