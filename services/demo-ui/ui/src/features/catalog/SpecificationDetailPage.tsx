import { useParams, Link } from 'react-router-dom';
import { ArrowLeft, Edit, Loader2, Tag, Clock, CheckCircle, XCircle, AlertCircle } from 'lucide-react';
import { useSpecification } from './api';
import './SpecificationListPage.css';

export default function SpecificationDetailPage() {
    const { id } = useParams<{ id: string }>();
    const { data: spec, isLoading, error } = useSpecification(id);

    if (isLoading) {
        return (
            <div className="specification-detail-page">
                <div className="loading-state">
                    <Loader2 className="spin" size={32} />
                    <p>Loading specification...</p>
                </div>
            </div>
        );
    }

    if (error || !spec) {
        return (
            <div className="specification-detail-page">
                <div className="error-state">
                    <AlertCircle size={48} className="error-icon" />
                    <p>{error ? `Error: ${error.message}` : 'Specification not found'}</p>
                    <Link to="/catalog/specifications" className="btn btn-primary">
                        Back to Specifications
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

    return (
        <div className="specification-detail-page">
            <div className="page-header">
                <div className="page-header-content">
                    <Link to="/catalog/specifications" className="back-link">
                        <ArrowLeft size={18} />
                        <span>Back to Specifications</span>
                    </Link>
                    <div className="header-title-area mt-2">
                        <div className="title-with-badge">
                            <h2>{spec.name}</h2>
                            {spec.isBundle && <span className="bundle-badge">Bundle</span>}
                        </div>
                        <p className="lead text-muted">{spec.productNumber}</p>
                    </div>
                </div>
                <Link to={`/catalog/specifications/${id}/edit`} className="btn btn-secondary">
                    <Edit size={18} />
                    <span>Edit Specification</span>
                </Link>
            </div>

            <div className="detail-grid mt-4">
                <div className="main-content">
                    <div className="card detail-section">
                        <h3>Overview</h3>
                        <div className="overview-content">
                            <p className="description-text">
                                {spec.description || <span className="text-muted">No description provided.</span>}
                            </p>
                            <div className="metadata-grid mt-4">
                                <div className="metadata-item">
                                    <span className="label">Lifecycle Status</span>
                                    <div className="value-with-icon">
                                        {getStatusIcon(spec.lifecycleStatus)}
                                        <span className={`status-badge ${spec.lifecycleStatus.toLowerCase()}`}>
                                            {spec.lifecycleStatus}
                                        </span>
                                    </div>
                                </div>
                                <div className="metadata-item">
                                    <span className="label">Last Updated</span>
                                    <span className="value">{new Date(spec.lastUpdate).toLocaleString()}</span>
                                </div>
                                {spec.validFor && (
                                    <div className="metadata-item">
                                        <span className="label">Validity Period</span>
                                        <span className="value">
                                            {spec.validFor.startDateTime ? new Date(spec.validFor.startDateTime).toLocaleDateString() : 'Start'}
                                            {' - '}
                                            {spec.validFor.endDateTime ? new Date(spec.validFor.endDateTime).toLocaleDateString() : 'End'}
                                        </span>
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>

                    <div className="card detail-section mt-4">
                        <div className="section-header">
                            <h3>Characteristics</h3>
                            <span className="badge badge-secondary">{Object.keys(spec.characteristics || {}).length} Items</span>
                        </div>
                        {Object.keys(spec.characteristics || {}).length === 0 ? (
                            <p className="empty-text">No characteristics defined for this specification.</p>
                        ) : (
                            <div className="characteristics-list mt-3">
                                {Object.entries(spec.characteristics || {}).map(([name, char]) => (
                                    <div key={name} className="char-item card">
                                        <div className="char-header">
                                            <div className="char-title">
                                                <Tag size={16} />
                                                <strong>{char.name}</strong>
                                                <span className="char-type-badge">{char.valueType}</span>
                                            </div>
                                            {char.configurable && <span className="badge badge-info-outline">Configurable</span>}
                                        </div>
                                        {char.description && (
                                            <p className="char-desc text-muted mt-2">{char.description}</p>
                                        )}
                                        {char.validValues && char.validValues.length > 0 && (
                                            <div className="char-values mt-2">
                                                <span className="label">Valid Values:</span>
                                                <div className="values-cloud">
                                                    {char.validValues.map(v => <span key={v} className="value-tag">{v}</span>)}
                                                </div>
                                            </div>
                                        )}
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                </div>

                <div className="sidebar-content">
                    <div className="card info-card">
                        <h3>Quick Stats</h3>
                        <div className="stats-list">
                            <div className="stat-item">
                                <span className="stat-label">Product SKU</span>
                                <span className="stat-value">{spec.productNumber}</span>
                            </div>
                            <div className="stat-item">
                                <span className="stat-label">Bundle</span>
                                <span className="stat-value">{spec.isBundle ? 'Yes' : 'No'}</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
