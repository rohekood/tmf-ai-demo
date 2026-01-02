import { useState, useCallback, type ReactNode } from 'react';
import './Toast.css';
import { NotificationContext } from './Toast';
import type { Toast, ToastType } from './Toast';
import { ToastItem } from './ToastItem';

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
                    <ToastItem key={toast.id} toast={toast} onClose={removeToast} />
                ))}
            </div>
        </NotificationContext.Provider>
    );
}
