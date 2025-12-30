import { useState, useCallback, type ReactNode } from 'react';
import { X, CheckCircle, AlertCircle, Info } from 'lucide-react';
import './Toast.css';
import { NotificationContext } from './Toast';
import type { Toast, ToastType } from './Toast';

export function NotificationProvider({ children }: { children: ReactNode }) {
    const [toasts, setToasts] = useState<Toast[]>([]);

    const removeToast = useCallback((id: string) => {
        setToasts((prev) => prev.filter((toast) => toast.id !== id));
    }, []);

    const showToast = useCallback((message: string, type: ToastType) => {
        const id = Math.random().toString(36).substring(2, 9);
        setToasts((prev) => [...prev, { id, message, type }]);

        setTimeout(() => {
            removeToast(id);
        }, 10000); // Auto-close after 10 seconds
    }, [removeToast]);

    return (
        <NotificationContext.Provider value={{ showToast }}>
            {children}
            <div className="toast-container">
                {toasts.map((toast) => (
                    <div key={toast.id} className={`toast toast--${toast.type}`} role="alert">
                        <div className="toast-icon">
                            {toast.type === 'success' && <CheckCircle size={20} />}
                            {toast.type === 'error' && <AlertCircle size={20} />}
                            {toast.type === 'info' && <Info size={20} />}
                        </div>
                        <p className="toast-message">{toast.message}</p>
                        <button className="toast-close" onClick={() => removeToast(toast.id)}>
                            <X size={16} />
                        </button>
                    </div>
                ))}
            </div>
        </NotificationContext.Provider>
    );
}
