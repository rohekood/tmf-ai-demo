import { useParams, Link } from 'react-router-dom';
import { ArrowLeft, Loader2 } from 'lucide-react';
import { useSpecification } from './api';
import SpecificationEditForm from './SpecificationEditForm.tsx';
import '../parties/PartyFormPage.css';

export default function SpecificationEditPage() {
    const { id } = useParams<{ id: string }>();
    const isNew = !id || id === 'new';
    const { data: spec, isLoading } = useSpecification(isNew ? undefined : id);

    if (!isNew && isLoading) {
        return (
            <div className="specification-edit-page">
                <div className="loading-state">
                    <Loader2 className="spin" size={32} />
                    <p>Loading specification...</p>
                </div>
            </div>
        );
    }

    if (!isNew && !spec) {
        return (
            <div className="specification-edit-page">
                <div className="error-state">
                    <p>Specification not found</p>
                    <Link to="/catalog/specifications" className="btn btn-primary">
                        Back to Specifications
                    </Link>
                </div>
            </div>
        );
    }

    return (
        <div className="specification-edit-page">
            <Link to={isNew ? "/catalog/specifications" : `/catalog/specifications/${id}`} className="back-link" style={{ marginBottom: '1rem', display: 'inline-flex' }}>
                <ArrowLeft size={18} />
                <span>Back to {isNew ? 'Specifications' : 'Details'}</span>
            </Link>

            <SpecificationEditForm specification={spec} isNew={isNew} />
        </div>
    );
}
