import { useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useSagaStatus } from './api';
import { PageLoader } from '../../design-system/components/common/PageLoader';
import OrderStepTracker from './OrderStepTracker';

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

    if (error) return <div className="p-8 text-red-600 text-center">Error fetching status.</div>;

    return (
        <div className="p-8 max-w-2xl mx-auto space-y-12">
            <div className="text-center">
                <h1 className="text-3xl font-bold">Processing Order</h1>
                <p className="text-gray-500 mt-2">Saga ID: {sagaId}</p>
            </div>

            <div className="bg-white p-8 rounded-lg shadow border border-gray-100">
                <OrderStepTracker sagaStatus={statusData?.status} />

                {statusData?.status === 'FAILED' && (
                    <div className="mt-8 p-4 bg-red-50 border border-red-200 rounded-md text-red-700">
                        <p className="font-bold">Order processing failed</p>
                        <p className="text-sm mt-1">{statusData.errorReason}</p>
                        <Link
                            to="/order/cart"
                            className="inline-block mt-3 text-sm font-medium text-red-700 underline hover:text-red-900"
                        >
                            Modify Cart
                        </Link>
                    </div>
                )}
            </div>
        </div>
    );
}
