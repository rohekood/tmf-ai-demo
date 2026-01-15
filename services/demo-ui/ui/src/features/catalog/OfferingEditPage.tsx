import { useParams, Link } from 'react-router-dom';
import { ArrowLeft, Loader2 } from 'lucide-react';
import { useOffering } from './api';
import OfferingEditForm from './OfferingEditForm';
import '../parties/PartyFormPage.css';

export default function OfferingEditPage() {
    const { id } = useParams<{ id: string }>();
    const isNew = !id || id === 'new';
    const { data: offering, isLoading } = useOffering(isNew ? undefined : id);

    if (!isNew && isLoading) {
        return (
            <div className="offering-edit-page">
                <div className="loading-state">
                    <Loader2 className="spin" size={32} />
                    <p>Loading offering...</p>
                </div>
            </div>
        );
    }

    if (!isNew && !offering) {
        return (
            <div className="offering-edit-page">
                <div className="error-state">
                    <p>Offering not found</p>
                    <Link to="/catalog/offerings" className="btn btn-primary">
                        Back to Offerings
                    </Link>
                </div>
            </div>
        );
    }

    return (
        <div className="offering-edit-page">
            <Link to={isNew ? "/catalog/offerings" : `/catalog/offerings/${id}`} className="back-link" style={{ marginBottom: '1rem', display: 'inline-flex' }}>
                <ArrowLeft size={18} />
                <span>Back to {isNew ? 'Offerings' : 'Details'}</span>
            </Link>

            <OfferingEditForm offering={offering} isNew={isNew} />
        </div>
    );
}
