import { useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useSagaStatus } from './api';
import { PageLoader } from '../../design-system/components/common/PageLoader';

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

    const getStepStatus = (stepName: string) => {
        if (!statusData) return 'pending';
        const s = statusData.status;

        switch (stepName) {
            case 'inventory':
                if (s === 'INVENTORY_RESERVED' || s === 'PAYMENT_AUTHORIZED' || s === 'COMPLETED') return 'completed';
                if (s === 'INVENTORY_FAILED') return 'failed';
                return s === 'STARTED' ? 'current' : 'pending';
            case 'payment':
                if (s === 'PAYMENT_AUTHORIZED' || s === 'COMPLETED') return 'completed';
                if (s === 'PAYMENT_FAILED') return 'failed';
                return s === 'INVENTORY_RESERVED' ? 'current' : 'pending';
            case 'order':
                if (s === 'COMPLETED') return 'completed';
                if (s === 'FAILED') return 'failed';
                return s === 'PAYMENT_AUTHORIZED' ? 'current' : 'pending';
            default:
                return 'pending';
        }
    };

    const steps = [
        { id: 'inventory', label: 'Inventory' },
        { id: 'payment', label: 'Payment' },
        { id: 'order', label: 'Order Created' }
    ];

    const stepColor = (status: string) => {
        if (status === 'completed') return 'bg-green-500 border-green-500';
        if (status === 'current') return 'bg-blue-500 border-blue-500 animate-pulse';
        if (status === 'failed') return 'bg-red-500 border-red-500';
        return 'bg-gray-200 border-gray-200';
    };

    return (
        <div className="p-8 max-w-2xl mx-auto space-y-12">
            <div className="text-center">
                <h1 className="text-3xl font-bold">Processing Order</h1>
                <p className="text-gray-500 mt-2">Saga ID: {sagaId}</p>
            </div>

            <div className="bg-white p-8 rounded-lg shadow border border-gray-100">
                <div className="space-y-8">
                    {steps.map((step) => {
                        const s = getStepStatus(step.id);
                        return (
                            <div key={step.id} className="flex items-center">
                                <div className={`flex-shrink-0 w-8 h-8 rounded-full border-2 flex items-center justify-center ${stepColor(s)}`}>
                                    {s === 'completed' && <span className="text-white font-bold">✓</span>}
                                    {s === 'failed' && <span className="text-white font-bold">✕</span>}
                                </div>
                                <div className="ml-4 flex-1">
                                    <h3 className={`text-lg font-medium ${s === 'current' ? 'text-blue-600' : 'text-gray-900'}`}>
                                        {step.label}
                                    </h3>
                                    {s === 'current' && <p className="text-sm text-gray-500">In progress...</p>}
                                </div>
                            </div>
                        );
                    })}
                </div>
                
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
