import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import QualifyPage from './QualifyPage';
import * as api from './api';

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

describe('QualifyPage', () => {
    beforeEach(() => {
        mockNavigate.mockClear();
    });

    it('renders qualification form with street, number, city, zip fields', () => {
        vi.mocked(api.useCheckQualification).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
            reset: vi.fn(),
        } as any);
        vi.mocked(api.useAddCartItem).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
        } as any);

        render(
            <MemoryRouter>
                <QualifyPage />
            </MemoryRouter>
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
        } as any);
        vi.mocked(api.useAddCartItem).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
        } as any);

        render(
            <MemoryRouter>
                <QualifyPage />
            </MemoryRouter>
        );

        // All fields start empty; verify required attribute is present
        const streetInput = screen.getByLabelText(/Street/i);
        expect(streetInput).toBeRequired();

        const numberInput = screen.getByLabelText(/Number/i);
        expect(numberInput).toBeRequired();

        const cityInput = screen.getByLabelText(/City/i);
        expect(cityInput).toBeRequired();

        const zipInput = screen.getByLabelText(/ZIP/i);
        expect(zipInput).toBeRequired();

        // Fill all fields and verify mutate is called on submit
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
            if (options?.onSuccess) options.onSuccess();
        });

        vi.mocked(api.useCheckQualification).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
            reset: vi.fn(),
            data: {
                sessionId: 'sess123',
                status: 'Qualified',
                qualifiedOffers: [
                    {
                        offeringId: 'off1',
                        offeringName: 'Super Fiber',
                        price: { amount: 49.99, currency: 'EUR', taxIncluded: true },
                        eligibility: 'QUALIFIED',
                    },
                ],
            },
        } as any);

        vi.mocked(api.useAddCartItem).mockReturnValue({
            mutate: mockAddToCart,
            isPending: false,
        } as any);

        render(
            <MemoryRouter>
                <QualifyPage />
            </MemoryRouter>
        );

        expect(screen.getByText('Super Fiber')).toBeInTheDocument();
        expect(screen.getByText(/49.99/)).toBeInTheDocument();
        expect(screen.getByText(/EUR/)).toBeInTheDocument();

        await user.click(screen.getByRole('button', { name: /Add to Cart/i }));
        expect(mockAddToCart).toHaveBeenCalled();
        expect(mockNavigate).toHaveBeenCalledWith('/order/cart');
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
        } as any);
        vi.mocked(api.useAddCartItem).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
        } as any);

        render(
            <MemoryRouter>
                <QualifyPage />
            </MemoryRouter>
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
        } as any);
        vi.mocked(api.useAddCartItem).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
        } as any);

        render(
            <MemoryRouter>
                <QualifyPage />
            </MemoryRouter>
        );

        expect(screen.getByText(/Failed to check qualification/i)).toBeInTheDocument();
        const retryButton = screen.getByRole('button', { name: /Retry/i });
        expect(retryButton).toBeInTheDocument();

        await user.click(retryButton);
        expect(mockReset).toHaveBeenCalled();
    });
});

