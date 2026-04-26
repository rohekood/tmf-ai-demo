import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import CartPage from './CartPage';
import * as api from './api';

vi.mock('./api', () => ({
    useCart: vi.fn(),
    useRemoveCartItem: vi.fn(),
}));

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom');
    return {
        ...actual,
        useNavigate: () => mockNavigate,
    };
});

describe('CartPage', () => {
    it('renders loading state', () => {
        vi.mocked(api.useCart).mockReturnValue({ isLoading: true } as any);
        vi.mocked(api.useRemoveCartItem).mockReturnValue({ mutate: vi.fn() } as any);

        render(
            <MemoryRouter>
                <CartPage />
            </MemoryRouter>
        );
        expect(screen.getByText('Loading...')).toBeInTheDocument(); // Loader
    });

    it('renders empty cart message', () => {
        vi.mocked(api.useCart).mockReturnValue({
            data: { items: [] },
            isLoading: false,
        } as any);

        render(
            <MemoryRouter>
                <CartPage />
            </MemoryRouter>
        );

        expect(screen.getByText('Your cart is empty.')).toBeInTheDocument();
    });

    it('renders cart items and handles checkout navigation', async () => {
        const user = userEvent.setup();
        vi.mocked(api.useCart).mockReturnValue({
            data: {
                id: 'c1',
                items: [{ id: 'i1', name: 'Fiber Plan', quantity: 1, price: 50, currency: 'EUR' }],
                totalPrice: 50,
                currency: 'EUR'
            },
            isLoading: false,
        } as any);

        render(
            <MemoryRouter>
                <CartPage />
            </MemoryRouter>
        );

        expect(screen.getByText('Fiber Plan')).toBeInTheDocument();
        
        // Test navigation
        await user.click(screen.getByRole('button', { name: /Proceed to Checkout/i }));
        expect(mockNavigate).toHaveBeenCalledWith('/order/checkout');
    });

    it('handles removing an item from the cart and reloads the page', async () => {
        const user = userEvent.setup();
        const mockRemoveMutate = vi.fn().mockImplementation((_, options) => {
            if (options?.onSuccess) options.onSuccess();
        });

        vi.mocked(api.useCart).mockReturnValue({
            data: {
                id: 'c1',
                items: [{ id: 'i1', name: 'Fiber Plan', quantity: 1, price: 50, currency: 'EUR' }],
                totalPrice: 50,
                currency: 'EUR'
            },
            isLoading: false,
        } as any);

        vi.mocked(api.useRemoveCartItem).mockReturnValue({
            mutate: mockRemoveMutate,
            isPending: false,
        } as any);

        const originalLocation = window.location;
        Object.defineProperty(window, 'location', {
            configurable: true,
            value: { reload: vi.fn() }
        });

        render(
            <MemoryRouter>
                <CartPage />
            </MemoryRouter>
        );

        await user.click(screen.getByRole('button', { name: /Remove/i }));
        
        expect(mockRemoveMutate).toHaveBeenCalledWith(
            { cartId: 'default-cart', itemId: 'i1' },
            expect.any(Object)
        );
        expect(window.location.reload).toHaveBeenCalled();

        Object.defineProperty(window, 'location', { configurable: true, value: originalLocation });
    });
    
    it('navigates to browse services when cart is empty', async () => {
        const user = userEvent.setup();
        vi.mocked(api.useCart).mockReturnValue({
            data: { items: [] },
            isLoading: false,
        } as any);

        render(
            <MemoryRouter>
                <CartPage />
            </MemoryRouter>
        );

        await user.click(screen.getByRole('button', { name: /Browse Services/i }));
        expect(mockNavigate).toHaveBeenCalledWith('/order/qualify');
    });
});
