import { useParams, Link } from 'react-router-dom';
import { ArrowLeft, Loader2 } from 'lucide-react';
import { useCatalog } from './api';
import CatalogEditForm from './CatalogEditForm';
import '../parties/PartyFormPage.css';

export default function CatalogEditPage() {
    const { id } = useParams<{ id: string }>();
    const isNew = !id || id === 'new';
    const { data: catalog, isLoading } = useCatalog(isNew ? undefined : id);

    if (!isNew && isLoading) {
        return (
            <div className="catalog-edit-page">
                <div className="loading-state">
                    <Loader2 className="spin" size={32} />
                    <p>Loading catalog...</p>
                </div>
            </div>
        );
    }

    if (!isNew && !catalog) {
        return (
            <div className="catalog-edit-page">
                <div className="error-state">
                    <p>Catalog not found</p>
                    <Link to="/catalog/catalogs" className="btn btn-primary">
                        Back to Catalogs
                    </Link>
                </div>
            </div>
        );
    }

    return (
        <div className="catalog-edit-page">
            <Link to={isNew ? "/catalog/catalogs" : `/catalog/catalogs/${id}`} className="back-link" style={{ marginBottom: '1rem', display: 'inline-flex' }}>
                <ArrowLeft size={18} />
                <span>Back to {isNew ? 'Catalogs' : 'Details'}</span>
            </Link>

            <CatalogEditForm catalog={catalog} isNew={isNew} />
        </div>
    );
}
