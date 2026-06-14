import { useParams, Link } from 'react-router-dom';
import { ArrowLeft, Edit, Loader2, Clock, CheckCircle, XCircle, AlertCircle } from 'lucide-react';
import { useCatalog } from './api';
import { formatDate, formatDateTime } from '../../lib/date';
import './CatalogDetail.css';

function getStatusIcon(status: string) {
    switch (status) {
        case 'Active': return <CheckCircle size={16} className="text-success" />;
        case 'Retired': return <XCircle size={16} className="text-danger" />;
        case 'Draft': return <Clock size={16} className="text-muted" />;
        default: return <AlertCircle size={16} />;
    }
}

export default function CatalogDetailPage() {
    const { id } = useParams<{ id: string }>();
    const { data: catalog, isLoading, error } = useCatalog(id);

    if (isLoading) {
        return (
            <div className="specification-detail-page">
                <div className="loading-state">
                    <Loader2 className="spin" size={32} />
                    <p>Loading catalog...</p>
                </div>
            </div>
        );
    }

    if (error || !catalog) {
        return (
            <div className="specification-detail-page">
                <div className="error-state">
                    <AlertCircle size={48} className="error-icon" />
                    <p>{error ? `Error: ${error.message}` : 'Catalog not found'}</p>
                    <Link to="/catalog/catalogs" className="btn btn-primary">
                        Back to Catalogs
                    </Link>
                </div>
            </div>
        );
    }

    return (
        <div className="specification-detail-page">
            <div className="page-header">
                <div className="page-header-content">
                    <Link to="/catalog/catalogs" className="back-link">
                        <ArrowLeft size={18} />
                        <span>Back to Catalogs</span>
                    </Link>
                    <div className="header-title-area mt-2">
                        <div className="title-with-badge">
                            <h2>{catalog.name}</h2>
                        </div>
                    </div>
                </div>
                <Link to={`/catalog/catalogs/${id}/edit`} className="btn btn-secondary">
                    <Edit size={18} />
                    <span>Edit Catalog</span>
                </Link>
            </div>

            <div className="detail-grid mt-4">
                <div className="main-content">
                    <div className="card detail-section">
                        <h3>Overview</h3>
                        <div className="overview-content">
                            <p className="description-text">
                                {catalog.description || <span className="text-muted">No description provided.</span>}
                            </p>
                            <div className="metadata-grid mt-4">
                                <div className="metadata-item">
                                    <span className="label">Lifecycle Status</span>
                                    <div className="value-with-icon">
                                        {getStatusIcon(catalog.lifecycleStatus)}
                                        <span className={`status-badge ${catalog.lifecycleStatus.toLowerCase()}`}>
                                            {catalog.lifecycleStatus}
                                        </span>
                                    </div>
                                </div>
                                <div className="metadata-item">
                                    <span className="label">Last Updated</span>
                                    <span className="value">{formatDateTime(catalog.lastUpdate)}</span>
                                </div>
                                {catalog.validFor && (
                                    <div className="metadata-item">
                                        <span className="label">Validity Period</span>
                                        <span className="value">
                                            {formatDate(catalog.validFor.startDateTime, 'Start')}
                                            {' - '}
                                            {formatDate(catalog.validFor.endDateTime, 'End')}
                                        </span>
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>
                </div>

                <div className="sidebar-content">
                    <div className="card info-card">
                        <h3>Quick Stats</h3>
                        <div className="stats-list">
                            <div className="stat-item">
                                <span className="stat-label">Status</span>
                                <span className="stat-value">{catalog.lifecycleStatus}</span>
                            </div>
                            <div className="stat-item">
                                <span className="stat-label">Last Updated</span>
                                <span className="stat-value">{formatDate(catalog.lastUpdate)}</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
