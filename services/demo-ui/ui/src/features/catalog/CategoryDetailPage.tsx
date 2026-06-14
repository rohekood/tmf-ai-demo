import { useParams, Link } from 'react-router-dom';
import { ArrowLeft, Edit, Loader2, Clock, CheckCircle, XCircle, AlertCircle, FolderTree } from 'lucide-react';
import { useCategory, useCategories } from './api';
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

export default function CategoryDetailPage() {
    const { id } = useParams<{ id: string }>();
    const { data: category, isLoading, error } = useCategory(id);
    const { data: allCategories = [] } = useCategories();

    if (isLoading) {
        return (
            <div className="specification-detail-page">
                <div className="loading-state">
                    <Loader2 className="spin" size={32} />
                    <p>Loading category...</p>
                </div>
            </div>
        );
    }

    if (error || !category) {
        return (
            <div className="specification-detail-page">
                <div className="error-state">
                    <AlertCircle size={48} className="error-icon" />
                    <p>{error ? `Error: ${error.message}` : 'Category not found'}</p>
                    <Link to="/catalog/categories" className="btn btn-primary">
                        Back to Categories
                    </Link>
                </div>
            </div>
        );
    }

    const parent = category.parentId
        ? allCategories.find(c => c.id === category.parentId)
        : undefined;

    return (
        <div className="specification-detail-page">
            <div className="page-header">
                <div className="page-header-content">
                    <Link to="/catalog/categories" className="back-link">
                        <ArrowLeft size={18} />
                        <span>Back to Categories</span>
                    </Link>
                    <div className="header-title-area mt-2">
                        <div className="title-with-badge">
                            <h2>{category.name}</h2>
                            {category.isRoot && <span className="bundle-badge">Root</span>}
                        </div>
                    </div>
                </div>
                <Link to={`/catalog/categories/${id}/edit`} className="btn btn-secondary">
                    <Edit size={18} />
                    <span>Edit Category</span>
                </Link>
            </div>

            <div className="detail-grid mt-4">
                <div className="main-content">
                    <div className="card detail-section">
                        <h3>Overview</h3>
                        <div className="overview-content">
                            <p className="description-text">
                                {category.description || <span className="text-muted">No description provided.</span>}
                            </p>
                            <div className="metadata-grid mt-4">
                                <div className="metadata-item">
                                    <span className="label">Lifecycle Status</span>
                                    <div className="value-with-icon">
                                        {getStatusIcon(category.lifecycleStatus)}
                                        <span className={`status-badge ${category.lifecycleStatus.toLowerCase()}`}>
                                            {category.lifecycleStatus}
                                        </span>
                                    </div>
                                </div>
                                <div className="metadata-item">
                                    <span className="label">Last Updated</span>
                                    <span className="value">{formatDateTime(category.lastUpdate)}</span>
                                </div>
                                {category.validFor && (
                                    <div className="metadata-item">
                                        <span className="label">Validity Period</span>
                                        <span className="value">
                                            {formatDate(category.validFor.startDateTime, 'Start')}
                                            {' - '}
                                            {formatDate(category.validFor.endDateTime, 'End')}
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
                                <span className="stat-label">Root Category</span>
                                <span className="stat-value">{category.isRoot ? 'Yes' : 'No'}</span>
                            </div>
                            <div className="stat-item">
                                <span className="stat-label">Status</span>
                                <span className="stat-value">{category.lifecycleStatus}</span>
                            </div>
                        </div>
                    </div>

                    {!category.isRoot && parent && (
                        <div className="card info-card">
                            <h3>Parent Category</h3>
                            <div className="stats-list">
                                <div className="stat-item">
                                    <span className="stat-label">
                                        <FolderTree size={14} /> {parent.name}
                                    </span>
                                </div>
                            </div>
                            <Link
                                to={`/catalog/categories/${parent.id}`}
                                className="btn btn-secondary mt-3"
                            >
                                View Parent
                            </Link>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}
