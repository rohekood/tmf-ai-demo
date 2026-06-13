import { useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useSagaStatus } from './api';
import { PageLoader } from '../../design-system/components/common/PageLoader';
import OrderStepTracker from './OrderStepTracker';
import './ordering.css';

export default function OrderStatusPage() {
    const { sagaId } = useParams<{ sagaId: string }>();
    const navigate = useNavigate();

    const { data: statusData, isLoading, error } = useSagaStatus(sagaId);

    useEffect(() => {
        if (statusData?.status === 'COMPLETED' && statusData.orderId) {
            // Wait 2 seconds to show the completed state, then redirect
            const t = setTimeout(() => {
                navigate(`/order/confirmation/${statusData.orderId}`);
            }, 2000);
            return () => clearTimeout(t);
        }
    }, [statusData, navigate]);

    if (isLoading && !statusData) return <PageLoader />;

    if (error) return <div className="page"><div className="alert alert-danger">Error fetching status.</div></div>;

    return (
        <div className="page page--narrow">
            <div style={{ textAlign: 'center' }}>
                <h1 className="page-title">Processing Order</h1>
                <p className="page-subtitle mono">Saga ID: {sagaId}</p>
            </div>

            <div className="card">
                <OrderStepTracker sagaStatus={statusData?.status} />

                {statusData?.status === 'FAILED' && (
                    <div className="alert alert-danger" style={{ marginTop: '1.5rem', display: 'block' }}>
                        <p style={{ fontWeight: 700 }}>Order processing failed</p>
                        <p style={{ fontSize: '0.875rem', marginTop: '0.25rem' }}>{statusData.errorReason}</p>
                        <Link to="/order/cart" className="btn-link" style={{ display: 'inline-block', marginTop: '0.75rem', color: 'inherit', textDecoration: 'underline' }}>
                            Modify Cart
                        </Link>
                    </div>
                )}
            </div>
        </div>
    );
}
