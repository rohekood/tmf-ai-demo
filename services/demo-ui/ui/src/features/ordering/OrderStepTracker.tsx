import type { SagaStatusResponse } from './types';

type StepId = 'inventory' | 'payment' | 'order';
type StepStatus = 'pending' | 'current' | 'completed' | 'failed';

const STEPS: { id: StepId; label: string }[] = [
    { id: 'inventory', label: 'Inventory' },
    { id: 'payment', label: 'Payment' },
    { id: 'order', label: 'Order Created' },
];

function getStepStatus(stepId: StepId, sagaStatus: SagaStatusResponse['status'] | undefined): StepStatus {
    if (!sagaStatus) return 'pending';

    switch (stepId) {
        case 'inventory':
            if (sagaStatus === 'INVENTORY_RESERVED' || sagaStatus === 'PAYMENT_AUTHORIZED' || sagaStatus === 'PAYMENT_FAILED' || sagaStatus === 'COMPLETED') return 'completed';
            if (sagaStatus === 'INVENTORY_FAILED') return 'failed';
            return sagaStatus === 'STARTED' ? 'current' : 'pending';
        case 'payment':
            if (sagaStatus === 'PAYMENT_AUTHORIZED' || sagaStatus === 'COMPLETED') return 'completed';
            if (sagaStatus === 'PAYMENT_FAILED') return 'failed';
            return sagaStatus === 'INVENTORY_RESERVED' ? 'current' : 'pending';
        case 'order':
            if (sagaStatus === 'COMPLETED') return 'completed';
            if (sagaStatus === 'FAILED') return 'failed';
            return sagaStatus === 'PAYMENT_AUTHORIZED' ? 'current' : 'pending';
        default:
            return 'pending';
    }
}

function stepColor(status: StepStatus): string {
    if (status === 'completed') return 'bg-green-500 border-green-500';
    if (status === 'current') return 'bg-blue-500 border-blue-500 animate-pulse';
    if (status === 'failed') return 'bg-red-500 border-red-500';
    return 'bg-gray-200 border-gray-200';
}

interface Props {
    sagaStatus?: SagaStatusResponse['status'];
}

export default function OrderStepTracker({ sagaStatus }: Props) {
    return (
        <div className="space-y-8">
            {STEPS.map((step) => {
                const status = getStepStatus(step.id, sagaStatus);
                return (
                    <div key={step.id} className="flex items-center">
                        <div
                            className={`flex-shrink-0 w-8 h-8 rounded-full border-2 flex items-center justify-center ${stepColor(status)}`}
                            data-testid={`step-${step.id}`}
                            data-status={status}
                        >
                            {status === 'completed' && <span className="text-white font-bold">✓</span>}
                            {status === 'failed' && <span className="text-white font-bold">✕</span>}
                        </div>
                        <div className="ml-4 flex-1">
                            <h3 className={`text-lg font-medium ${status === 'current' ? 'text-blue-600' : 'text-gray-900'}`}>
                                {step.label}
                            </h3>
                            {status === 'current' && <p className="text-sm text-gray-500">In progress...</p>}
                        </div>
                    </div>
                );
            })}
        </div>
    );
}
