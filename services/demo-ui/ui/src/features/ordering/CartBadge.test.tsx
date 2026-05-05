import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { CartBadge } from './CartBadge';
import * as api from './api';

vi.mock('./api', () => ({
    useCart: vi.fn(),
}));

function withQueryClient(ui: React.ReactElement) {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return <QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>;
}

beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
});

describe('CartBadge', () => {
    it('renders nothing when cart is empty', () => {
        vi.mocked(api.useCart).mockReturnValue({ data: { items: [] } } as any);
        localStorage.setItem('cartId', 'cart-1');

        const { container } = render(withQueryClient(<CartBadge />));
        expect(container.firstChild).toBeNull();
    });

    it('renders nothing when there is no cartId in localStorage', () => {
        vi.mocked(api.useCart).mockReturnValue({ data: undefined } as any);

        const { container } = render(withQueryClient(<CartBadge />));
        expect(container.firstChild).toBeNull();
    });

    it('renders item count badge when cart has items', () => {
        vi.mocked(api.useCart).mockReturnValue({
            data: {
                items: [
                    { id: 'i1', offeringId: 'o1', quantity: 1, price: 10, currency: 'EUR' },
                    { id: 'i2', offeringId: 'o2', quantity: 2, price: 20, currency: 'EUR' },
                ],
            },
        } as any);
        localStorage.setItem('cartId', 'cart-1');

        render(withQueryClient(<CartBadge />));

        const badge = screen.getByLabelText('2 items in cart');
        expect(badge).toBeInTheDocument();
        expect(badge).toHaveTextContent('2');
    });

    it('uses singular label for one item', () => {
        vi.mocked(api.useCart).mockReturnValue({
            data: {
                items: [{ id: 'i1', offeringId: 'o1', quantity: 1, price: 10, currency: 'EUR' }],
            },
        } as any);
        localStorage.setItem('cartId', 'cart-1');

        render(withQueryClient(<CartBadge />));

        expect(screen.getByLabelText('1 item in cart')).toBeInTheDocument();
    });

    it('reads cartId from localStorage and passes it to useCart', () => {
        vi.mocked(api.useCart).mockReturnValue({ data: undefined } as any);
        localStorage.setItem('cartId', 'my-cart-id');

        render(withQueryClient(<CartBadge />));

        expect(api.useCart).toHaveBeenCalledWith('my-cart-id');
    });
});
