import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect } from 'vitest';
import { NotificationProvider, useNotification } from './Toast';

const TestComponent = () => {
    const { showToast } = useNotification();
    return (
        <div>
            <button onClick={() => showToast('Success Message', 'success')}>Show Success</button>
            <button onClick={() => showToast('Error Message', 'error')}>Show Error</button>
        </div>
    );
};

describe('Toast Component', () => {

    it('shows toast when function is called', async () => {
        const user = userEvent.setup();
        render(
            <NotificationProvider>
                <TestComponent />
            </NotificationProvider>
        );

        await user.click(screen.getByRole('button', { name: /show success/i }));

        const toast = await screen.findByRole('alert');
        expect(toast).toHaveTextContent('Success Message');
        expect(toast).toHaveClass('toast--success');
    });

    it('shows error toast', async () => {
        const user = userEvent.setup();
        render(
            <NotificationProvider>
                <TestComponent />
            </NotificationProvider>
        );

        await user.click(screen.getByRole('button', { name: /show error/i }));

        const toast = await screen.findByRole('alert');
        expect(toast).toHaveTextContent('Error Message');
        expect(toast).toHaveClass('toast--error');
    });

    it('closes toast when close button is clicked', async () => {
        const user = userEvent.setup();

        render(
            <NotificationProvider>
                <TestComponent />
            </NotificationProvider>
        );

        await user.click(screen.getByRole('button', { name: /show success/i }));

        const toast = await screen.findByRole('alert');
        expect(toast).toBeInTheDocument();

        // Click the close button
        const closeButton = toast.querySelector('.toast-close');
        expect(closeButton).toBeInTheDocument();
        await user.click(closeButton!);

        // Toast should be gone
        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    });
});
