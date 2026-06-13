import type { SagaStatusResponse } from './types';
import './ordering.css';

type StepId = 'inventory' | 'payment' | 'order';
type StepStatus = 'pending' | 'current' | 'completed' | 'failed';

const STEPS: { id: StepId; label: string }[] = [
    { id: 'inventory', label: 'Inventory' },
    { id: 'payment', label: 'Payment' },
    { id: 'order', label: 'Order Created' },
];

const INVENTORY_COMPLETED_STATES = ['INVENTORY_RESERVED', 'PAYMENT_AUTHORIZED', 'PAYMENT_FAILED', 'COMPLETED'] as const;
const PAYMENT_COMPLETED_STATES = ['PAYMENT_AUTHORIZED', 'COMPLETED'] as const;

function getStepStatus(stepId: StepId, sagaStatus: SagaStatusResponse['status'] | undefined): StepStatus {
    if (!sagaStatus) return 'pending';

    switch (stepId) {
        case 'inventory':
            if ((INVENTORY_COMPLETED_STATES as readonly string[]).includes(sagaStatus)) return 'completed';
            if (sagaStatus === 'INVENTORY_FAILED') return 'failed';
            return sagaStatus === 'STARTED' ? 'current' : 'pending';
        case 'payment':
            if ((PAYMENT_COMPLETED_STATES as readonly string[]).includes(sagaStatus)) return 'completed';
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

interface Props {
    sagaStatus?: SagaStatusResponse['status'];
}

export default function OrderStepTracker({ sagaStatus }: Props) {
    return (
        <div className="step-tracker">
            {STEPS.map((step) => {
                const status = getStepStatus(step.id, sagaStatus);
                return (
                    <div key={step.id} className="step">
                        <div
                            className={`step__marker step__marker--${status}`}
                            data-testid={`step-${step.id}`}
                            data-status={status}
                        >
                            {status === 'completed' && <span>✓</span>}
                            {status === 'failed' && <span>✕</span>}
                        </div>
                        <div className="spacer">
                            <h3 className={`step__label ${status === 'current' ? 'step__label--current' : ''}`}>
                                {step.label}
                            </h3>
                            {status === 'current' && <p className="step__hint">In progress...</p>}
                        </div>
                    </div>
                );
            })}
        </div>
    );
}
