import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import CheckoutPage from './CheckoutPage';
import * as api from './api';

vi.mock('./api', () => ({
    useCheckout: vi.fn(),
    useCart: vi.fn(),
}));

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom');
    return {
        ...actual,
        useNavigate: () => mockNavigate,
    };
});

describe('CheckoutPage', () => {
    it('renders checkout form', async () => {
        const user = userEvent.setup();
        vi.mocked(api.useCheckout).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
        } as any);
        vi.mocked(api.useCart).mockReturnValue({
            data: { items: [] },
            isLoading: false,
        } as any);

        render(
            <MemoryRouter>
                <CheckoutPage />
            </MemoryRouter>
        );

        expect(screen.getByText('Checkout')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Place Order/i })).toBeInTheDocument();
        
        // Test changing payment method
        const paypalRadio = screen.getByRole('radio', { name: /PayPal/i });
        await user.click(paypalRadio);
        expect(paypalRadio).toBeChecked();
    });

    it('submits order and navigates to status page', async () => {
        const user = userEvent.setup();
        const mockMutate = vi.fn().mockImplementation((_, options) => {
            options.onSuccess({ sagaId: 'saga123' });
        });

        vi.mocked(api.useCheckout).mockReturnValue({
            mutate: mockMutate,
            isPending: false,
        } as any);
        vi.mocked(api.useCart).mockReturnValue({
            data: { items: [{ id: '1', quantity: 1, price: 10, currency: 'EUR' }] },
            isLoading: false,
        } as any);

        render(
            <MemoryRouter>
                <CheckoutPage />
            </MemoryRouter>
        );

        await user.click(screen.getByRole('button', { name: /Place Order/i }));
        
        expect(mockMutate).toHaveBeenCalled();
        expect(mockNavigate).toHaveBeenCalledWith('/order/status/saga123');
    });
});
