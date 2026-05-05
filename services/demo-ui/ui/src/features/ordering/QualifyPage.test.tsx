import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import QualifyPage from './QualifyPage';
import * as api from './api';
import { NotificationProvider } from '../../design-system/components/common/ToastProvider';

vi.mock('./api', () => ({
    useCheckQualification: vi.fn(),
    useAddCartItem: vi.fn(),
}));

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom');
    return {
        ...actual,
        useNavigate: () => mockNavigate,
    };
});

const sessionWithOfferings = {
    id: 'sess123',
    qualifiedOfferings: [{ offeringId: 'off1', name: 'Super Fiber', price: 50, currency: 'EUR' }]
};

beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
});

describe('QualifyPage', () => {
    it('renders qualification form and submits with typed address values', async () => {
        const user = userEvent.setup();
        const mockCheckQualify = vi.fn();
        vi.mocked(api.useCheckQualification).mockReturnValue({
            mutate: mockCheckQualify,
            isPending: false,
        } as any);
        vi.mocked(api.useAddCartItem).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
        } as any);

        render(
            <NotificationProvider>
                <MemoryRouter>
                    <QualifyPage />
                </MemoryRouter>
            </NotificationProvider>
        );

        expect(screen.getByText('Service Qualification')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Check Availability/i })).toBeInTheDocument();
        
        // Test all inputs update correctly
        const [streetInput, cityInput, postcodeInput, countryInput] = screen.getAllByRole('textbox');
        await user.clear(streetInput);
        await user.type(streetInput, 'New Street');
        expect(streetInput).toHaveValue('New Street');

        await user.clear(cityInput);
        await user.type(cityInput, 'Helsinki');
        expect(cityInput).toHaveValue('Helsinki');

        await user.clear(postcodeInput);
        await user.type(postcodeInput, '00100');
        expect(postcodeInput).toHaveValue('00100');

        await user.clear(countryInput);
        await user.type(countryInput, 'FI');
        expect(countryInput).toHaveValue('FI');

        // Submit and verify mutation called with the new address values
        await user.click(screen.getByRole('button', { name: /Check Availability/i }));
        expect(mockCheckQualify).toHaveBeenCalledWith({
            address: { street: 'New Street', city: 'Helsinki', postcode: '00100', country: 'FI' }
        });
    });

    it('submits form, shows offerings, and navigates on add to cart', async () => {
        const user = userEvent.setup();
        const mockCheckQualify = vi.fn();
        const mockAddToCart = vi.fn().mockImplementation((_, options) => {
            if (options?.onSuccess) options.onSuccess({ cartId: 'cart-abc', items: [], totalPrice: 0, currency: 'EUR' });
        });

        vi.mocked(api.useCheckQualification).mockReturnValue({
            mutate: mockCheckQualify,
            isPending: false,
            data: sessionWithOfferings
        } as any);
        
        vi.mocked(api.useAddCartItem).mockReturnValue({
            mutate: mockAddToCart,
            isPending: false,
        } as any);

        render(
            <NotificationProvider>
                <MemoryRouter>
                    <QualifyPage />
                </MemoryRouter>
            </NotificationProvider>
        );

        await user.click(screen.getByRole('button', { name: /Check Availability/i }));
        expect(mockCheckQualify).toHaveBeenCalled();

        // Should render the offering and "Add to Cart" button
        expect(screen.getByText('Super Fiber')).toBeInTheDocument();
        await user.click(screen.getByRole('button', { name: /Add to Cart/i }));

        expect(mockAddToCart).toHaveBeenCalled();
        expect(mockNavigate).toHaveBeenCalledWith('/order/cart');
    });

    it('saves cartId to localStorage on add item success', async () => {
        const user = userEvent.setup();
        const mockAddToCart = vi.fn().mockImplementation((_, options) => {
            if (options?.onSuccess) options.onSuccess({ cartId: 'returned-cart-id', items: [], totalPrice: 0, currency: 'EUR' });
        });

        vi.mocked(api.useCheckQualification).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
            data: sessionWithOfferings
        } as any);

        vi.mocked(api.useAddCartItem).mockReturnValue({
            mutate: mockAddToCart,
            isPending: false,
        } as any);

        render(
            <NotificationProvider>
                <MemoryRouter>
                    <QualifyPage />
                </MemoryRouter>
            </NotificationProvider>
        );

        await user.click(screen.getByRole('button', { name: /Add to Cart/i }));

        expect(localStorage.getItem('cartId')).toBe('returned-cart-id');
    });

    it('shows session expired banner when 422 session expired error is returned', async () => {
        const user = userEvent.setup();
        const axiosError = {
            isAxiosError: true,
            response: { status: 422, data: { error: 'session expired' } },
        };

        const mockAddToCart = vi.fn().mockImplementation((_, options) => {
            if (options?.onError) options.onError(axiosError);
        });

        vi.mocked(api.useCheckQualification).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
            data: sessionWithOfferings
        } as any);

        vi.mocked(api.useAddCartItem).mockReturnValue({
            mutate: mockAddToCart,
            isPending: false,
        } as any);

        render(
            <NotificationProvider>
                <MemoryRouter>
                    <QualifyPage />
                </MemoryRouter>
            </NotificationProvider>
        );

        await user.click(screen.getByRole('button', { name: /Add to Cart/i }));

        expect(screen.getByRole('alert')).toHaveTextContent(/session expired/i);
    });

    it('shows not eligible toast when 422 not eligible error is returned', async () => {
        const user = userEvent.setup();
        const axiosError = {
            isAxiosError: true,
            response: { status: 422, data: { error: 'offering not eligible' } },
        };

        const mockAddToCart = vi.fn().mockImplementation((_, options) => {
            if (options?.onError) options.onError(axiosError);
        });

        vi.mocked(api.useCheckQualification).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
            data: sessionWithOfferings
        } as any);

        vi.mocked(api.useAddCartItem).mockReturnValue({
            mutate: mockAddToCart,
            isPending: false,
        } as any);

        render(
            <NotificationProvider>
                <MemoryRouter>
                    <QualifyPage />
                </MemoryRouter>
            </NotificationProvider>
        );

        await user.click(screen.getByRole('button', { name: /Add to Cart/i }));

        const toast = await screen.findByRole('alert');
        expect(toast).toHaveTextContent('Not available at your address');
    });

    it('shows generic error toast for non-422 errors', async () => {
        const user = userEvent.setup();
        const genericError = new Error('Network error');

        const mockAddToCart = vi.fn().mockImplementation((_, options) => {
            if (options?.onError) options.onError(genericError);
        });

        vi.mocked(api.useCheckQualification).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
            data: sessionWithOfferings
        } as any);

        vi.mocked(api.useAddCartItem).mockReturnValue({
            mutate: mockAddToCart,
            isPending: false,
        } as any);

        render(
            <NotificationProvider>
                <MemoryRouter>
                    <QualifyPage />
                </MemoryRouter>
            </NotificationProvider>
        );

        await user.click(screen.getByRole('button', { name: /Add to Cart/i }));

        const toast = await screen.findByRole('alert');
        expect(toast).toHaveTextContent('Failed to add item – try again');
    });

    it('dismisses session expired banner when dismiss button is clicked', async () => {
        const user = userEvent.setup();
        const axiosError = {
            isAxiosError: true,
            response: { status: 422, data: { error: 'session expired' } },
        };

        const mockAddToCart = vi.fn().mockImplementation((_, options) => {
            if (options?.onError) options.onError(axiosError);
        });

        vi.mocked(api.useCheckQualification).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
            data: sessionWithOfferings
        } as any);

        vi.mocked(api.useAddCartItem).mockReturnValue({
            mutate: mockAddToCart,
            isPending: false,
        } as any);

        render(
            <NotificationProvider>
                <MemoryRouter>
                    <QualifyPage />
                </MemoryRouter>
            </NotificationProvider>
        );

        // Trigger the session expired banner
        await user.click(screen.getByRole('button', { name: /Add to Cart/i }));
        expect(screen.getByRole('alert')).toBeInTheDocument();

        // Dismiss it
        await user.click(screen.getByRole('button', { name: /Dismiss/i }));
        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    });
});
