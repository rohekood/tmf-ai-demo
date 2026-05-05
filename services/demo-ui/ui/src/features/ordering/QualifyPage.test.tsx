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
    sessionId: 'sess123',
    status: 'Qualified' as const,
    qualifiedOffers: [{
        offeringId: 'off1',
        offeringName: 'Super Fiber',
        price: { amount: 49.99, currency: 'EUR', taxIncluded: true },
        eligibility: 'QUALIFIED',
    }],
};

beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
});

describe('QualifyPage', () => {
    it('renders qualification form with street, number, city, zip fields', () => {
        vi.mocked(api.useCheckQualification).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
            reset: vi.fn(),
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
        vi.mocked(api.useAddCartItem).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(
            <NotificationProvider>
                <MemoryRouter>
                    <QualifyPage />
                </MemoryRouter>
            </NotificationProvider>
        );

        expect(screen.getByText('Service Qualification')).toBeInTheDocument();
        expect(screen.getByLabelText(/Street/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/Number/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/City/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/ZIP/i)).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Check Availability/i })).toBeInTheDocument();
    });

    it('form validation: all fields are required', async () => {
        const user = userEvent.setup();
        const mockCheckQualify = vi.fn();
        vi.mocked(api.useCheckQualification).mockReturnValue({
            mutate: mockCheckQualify,
            isPending: false,
            reset: vi.fn(),
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
        vi.mocked(api.useAddCartItem).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(
            <NotificationProvider>
                <MemoryRouter>
                    <QualifyPage />
                </MemoryRouter>
            </NotificationProvider>
        );

        const streetInput = screen.getByLabelText(/Street/i);
        expect(streetInput).toBeRequired();

        const numberInput = screen.getByLabelText(/Number/i);
        expect(numberInput).toBeRequired();

        const cityInput = screen.getByLabelText(/City/i);
        expect(cityInput).toBeRequired();

        const zipInput = screen.getByLabelText(/ZIP/i);
        expect(zipInput).toBeRequired();

        await user.type(streetInput, 'Main St');
        await user.type(numberInput, '10');
        await user.type(cityInput, 'Tallinn');
        await user.type(zipInput, '10001');
        await user.click(screen.getByRole('button', { name: /Check Availability/i }));
        expect(mockCheckQualify).toHaveBeenCalledWith(
            { address: { street: 'Main St', number: '10', city: 'Tallinn', zip: '10001' } }
        );
    });

    it('on success (Qualified): renders OfferingCard list with offeringName, price, currency, Add to Cart', async () => {
        const user = userEvent.setup();
        const mockAddToCart = vi.fn().mockImplementation((_, options) => {
            if (options?.onSuccess) options.onSuccess({ cartId: 'cart-abc' });
        });

        vi.mocked(api.useCheckQualification).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
            reset: vi.fn(),
            data: sessionWithOfferings,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        vi.mocked(api.useAddCartItem).mockReturnValue({
            mutate: mockAddToCart,
            isPending: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(
            <NotificationProvider>
                <MemoryRouter>
                    <QualifyPage />
                </MemoryRouter>
            </NotificationProvider>
        );

        expect(screen.getByText('Super Fiber')).toBeInTheDocument();
        expect(screen.getByText(/49.99/)).toBeInTheDocument();
        expect(screen.getByText(/EUR/)).toBeInTheDocument();

        await user.click(screen.getByRole('button', { name: /Add to Cart/i }));
        expect(mockAddToCart).toHaveBeenCalled();
        expect(mockNavigate).toHaveBeenCalledWith('/order/cart');
    });

    it('saves cartId to localStorage on add item success', async () => {
        const user = userEvent.setup();
        const mockAddToCart = vi.fn().mockImplementation((_, options) => {
            if (options?.onSuccess) options.onSuccess({ cartId: 'returned-cart-id' });
        });

        vi.mocked(api.useCheckQualification).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
            reset: vi.fn(),
            data: sessionWithOfferings,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        vi.mocked(api.useAddCartItem).mockReturnValue({
            mutate: mockAddToCart,
            isPending: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

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

    it('shows session expired banner when 422 SESSION_EXPIRED error is returned', async () => {
        const user = userEvent.setup();
        const axiosError = {
            isAxiosError: true,
            response: { status: 422, data: { error: 'SESSION_EXPIRED' } },
        };

        const mockAddToCart = vi.fn().mockImplementation((_, options) => {
            if (options?.onError) options.onError(axiosError);
        });

        vi.mocked(api.useCheckQualification).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
            reset: vi.fn(),
            data: sessionWithOfferings,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        vi.mocked(api.useAddCartItem).mockReturnValue({
            mutate: mockAddToCart,
            isPending: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

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
            reset: vi.fn(),
            data: sessionWithOfferings,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        vi.mocked(api.useAddCartItem).mockReturnValue({
            mutate: mockAddToCart,
            isPending: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

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
            reset: vi.fn(),
            data: sessionWithOfferings,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        vi.mocked(api.useAddCartItem).mockReturnValue({
            mutate: mockAddToCart,
            isPending: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

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
            response: { status: 422, data: { error: 'SESSION_EXPIRED' } },
        };

        const mockAddToCart = vi.fn().mockImplementation((_, options) => {
            if (options?.onError) options.onError(axiosError);
        });

        vi.mocked(api.useCheckQualification).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
            reset: vi.fn(),
            data: sessionWithOfferings,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        vi.mocked(api.useAddCartItem).mockReturnValue({
            mutate: mockAddToCart,
            isPending: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(
            <NotificationProvider>
                <MemoryRouter>
                    <QualifyPage />
                </MemoryRouter>
            </NotificationProvider>
        );

        await user.click(screen.getByRole('button', { name: /Add to Cart/i }));
        expect(screen.getByRole('alert')).toBeInTheDocument();

        await user.click(screen.getByRole('button', { name: /Dismiss/i }));
        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    });

    it('re-check availability button triggers re-qualification', async () => {
        const user = userEvent.setup();
        const axiosError = {
            isAxiosError: true,
            response: { status: 422, data: { error: 'SESSION_EXPIRED' } },
        };

        const mockAddToCart = vi.fn().mockImplementation((_, options) => {
            if (options?.onError) options.onError(axiosError);
        });
        const mockCheckQualify = vi.fn();
        const mockReset = vi.fn();

        vi.mocked(api.useCheckQualification).mockReturnValue({
            mutate: mockCheckQualify,
            isPending: false,
            reset: mockReset,
            data: sessionWithOfferings,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        vi.mocked(api.useAddCartItem).mockReturnValue({
            mutate: mockAddToCart,
            isPending: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(
            <NotificationProvider>
                <MemoryRouter>
                    <QualifyPage />
                </MemoryRouter>
            </NotificationProvider>
        );

        await user.click(screen.getByRole('button', { name: /Add to Cart/i }));
        expect(screen.getByRole('alert')).toBeInTheDocument();

        await user.click(screen.getByRole('button', { name: /re-check availability/i }));
        expect(mockReset).toHaveBeenCalled();
        expect(mockCheckQualify).toHaveBeenCalled();
        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    });

    it('on Unqualified: shows reason message and re-check button', () => {
        const mockReset = vi.fn();
        vi.mocked(api.useCheckQualification).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
            reset: mockReset,
            data: {
                sessionId: '',
                status: 'Unqualified',
                qualifiedOffers: [],
                unavailabilityReason: 'Outside service area',
            },
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
        vi.mocked(api.useAddCartItem).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(
            <NotificationProvider>
                <MemoryRouter>
                    <QualifyPage />
                </MemoryRouter>
            </NotificationProvider>
        );

        expect(screen.getByText('Service Not Available')).toBeInTheDocument();
        expect(screen.getByText('Outside service area')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Check Another Address/i })).toBeInTheDocument();
    });

    it('on error: shows error message and retry button', async () => {
        const user = userEvent.setup();
        const mockReset = vi.fn();
        vi.mocked(api.useCheckQualification).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
            reset: mockReset,
            error: new Error('Network error'),
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
        vi.mocked(api.useAddCartItem).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(
            <NotificationProvider>
                <MemoryRouter>
                    <QualifyPage />
                </MemoryRouter>
            </NotificationProvider>
        );

        expect(screen.getByText(/Failed to check qualification/i)).toBeInTheDocument();
        const retryButton = screen.getByRole('button', { name: /Retry/i });
        expect(retryButton).toBeInTheDocument();

        await user.click(retryButton);
        expect(mockReset).toHaveBeenCalled();
    });
});
