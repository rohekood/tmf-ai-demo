import { useParams, Link } from 'react-router-dom';
import './ordering.css';

export default function OrderConfirmationPage() {
    const { orderId } = useParams<{ orderId: string }>();

    return (
        <div className="page page--narrow confirm">
            <div className="confirm__check">✓</div>
            <h1 className="page-title" style={{ fontSize: '2.25rem' }}>Order Confirmed!</h1>
            <p className="page-subtitle" style={{ fontSize: '1.125rem' }}>Thank you for your purchase.</p>

            <div className="card confirm__id-card">
                <p className="muted" style={{ marginBottom: '0.35rem' }}>Your Order ID is</p>
                <p className="confirm__id">{orderId}</p>
            </div>

            <div style={{ paddingTop: '1.5rem' }}>
                <Link to="/" className="btn-link">&larr; Return to Dashboard</Link>
            </div>
        </div>
    );
}
