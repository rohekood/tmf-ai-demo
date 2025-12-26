import { useParams, Link } from 'react-router-dom';
import { ArrowLeft, Loader2 } from 'lucide-react';
import { useCustomer } from './api';
import CustomerEditForm from './CustomerEditForm';
import '../parties/PartyFormPage.css';

export default function CustomerEditPage() {
    const { id } = useParams<{ id: string }>();
    const { data: customer, isLoading } = useCustomer(id);

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

    if (!customer) {
        return (
            <div className="customer-edit-page">
                <div className="error-state">
                    <p>Customer not found</p>
                    <Link to="/customers" className="btn btn-primary">
                        Back to Customers
                    </Link>
                </div>
            </div>
        );
    }

    return (
        <div className="customer-edit-page">
            <Link to={`/customers/${id}`} className="back-link" style={{ marginBottom: '1rem', display: 'inline-flex' }}>
                <ArrowLeft size={18} />
                <span>Back to Details</span>
            </Link>

            <CustomerEditForm customer={customer} />
        </div>
    );
}
