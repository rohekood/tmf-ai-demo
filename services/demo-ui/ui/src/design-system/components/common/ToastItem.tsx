import { X, CheckCircle, AlertCircle, Info } from 'lucide-react';
import type { Toast } from './Toast';
import './Toast.css';

interface ToastItemProps {
    toast: Toast;
    onClose: (id: string) => void;
}

export function ToastItem({ toast, onClose }: ToastItemProps) {
    return (
        <div className={`toast toast--${toast.type}`} role="alert">
            <div className="toast-icon">
                {toast.type === 'success' && <CheckCircle size={20} />}
                {toast.type === 'error' && <AlertCircle size={20} />}
                {toast.type === 'info' && <Info size={20} />}
            </div>
            <p className="toast-message">{toast.message}</p>
            <button className="toast-close" onClick={() => onClose(toast.id)}>
                <X size={16} />
            </button>
        </div>
    );
}
