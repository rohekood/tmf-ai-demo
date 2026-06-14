import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import OfferingDetailPage from './OfferingDetailPage';
import * as api from './api';
import type { LifecycleStatus, ProductOffering } from './types';

const queryClient = new QueryClient({
    defaultOptions: {
        queries: { retry: false },
    },
});

vi.mock('./api', async () => {
    const actual = await vi.importActual('./api');
    return {
        ...actual,
        useOffering: vi.fn(),
    };
});

const renderComponent = (initialEntries = ['/catalog/offerings/off-1']) => {
    return render(
        <QueryClientProvider client={queryClient}>
            <MemoryRouter initialEntries={initialEntries}>
                <Routes>
                    <Route path="/catalog/offerings/:id" element={<OfferingDetailPage />} />
                </Routes>
            </MemoryRouter>
        </QueryClientProvider>
    );
};

const sampleOffering: ProductOffering = {
    id: 'off-1',
    name: 'Premium Fiber Plan',
    description: 'High speed fiber',
    lifecycleStatus: 'Active',
    validFor: { startDateTime: '2026-01-01T00:00:00Z' },
    lastUpdate: '2026-06-01T00:00:00Z',
    isBundle: false,
    isSellable: true,
    productOfferingPrice: [
        { id: 'p1', priceType: 'recurring', price: { unit: 'USD', value: 49.99 }, unitOfMeasure: 'month' },
    ],
    categories: [{ id: 'c1', name: 'Internet', isRoot: true, validFor: {}, lastUpdate: '', lifecycleStatus: 'Active' }],
};

describe('OfferingDetailPage', () => {
    it('renders the offering display view with name, status and pricing', () => {
        vi.mocked(api.useOffering).mockReturnValue({
            data: sampleOffering,
            isLoading: false,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();

        expect(screen.getByRole('heading', { name: 'Premium Fiber Plan' })).toBeInTheDocument();
        expect(screen.getByText('Active')).toBeInTheDocument();
        expect(screen.getByText('Recurring')).toBeInTheDocument();
        expect(screen.getByText('Internet')).toBeInTheDocument();
        expect(screen.getByRole('link', { name: /edit offering/i })).toHaveAttribute(
            'href',
            '/catalog/offerings/off-1/edit'
        );
    });

    it('shows a loading state while fetching', () => {
        vi.mocked(api.useOffering).mockReturnValue({
            data: undefined,
            isLoading: true,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();

        expect(screen.getByText('Loading offering...')).toBeInTheDocument();
    });

    it('shows a not-found state when the offering is missing', () => {
        vi.mocked(api.useOffering).mockReturnValue({
            data: undefined,
            isLoading: false,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();

        expect(screen.getByText('Offering not found')).toBeInTheDocument();
    });

    it('shows an error message when the query fails', () => {
        vi.mocked(api.useOffering).mockReturnValue({
            data: undefined,
            isLoading: false,
            error: new Error('explode'),
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();

        expect(screen.getByText('Error: explode')).toBeInTheDocument();
    });

    it.each(['Active', 'Retired', 'Draft', 'Suspended'] as LifecycleStatus[])(
        'renders the %s status icon branch',
        (status) => {
            vi.mocked(api.useOffering).mockReturnValue({
                data: { ...sampleOffering, lifecycleStatus: status },
                isLoading: false,
                error: null,
            } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

            renderComponent();

            expect(screen.getAllByText(status).length).toBeGreaterThan(0);
        }
    );

    it('formats a EUR price with the euro symbol and no stray dollar sign', () => {
        vi.mocked(api.useOffering).mockReturnValue({
            data: {
                ...sampleOffering,
                productOfferingPrice: [
                    { id: 'p2', priceType: 'one_time', price: { unit: 'EUR', value: 10 } },
                ],
            },
            isLoading: false,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();

        const price = screen.getByText(/10\.00/);
        expect(price.textContent).toContain('€');
        expect(price.textContent).not.toContain('$');
        expect(screen.getByText('One-time')).toBeInTheDocument();
    });

    it('renders price alterations (discount and fee) and an open-ended validity', () => {
        vi.mocked(api.useOffering).mockReturnValue({
            data: {
                ...sampleOffering,
                isSellable: false,
                validFor: { endDateTime: '2027-01-01T00:00:00Z' },
                productOfferingPrice: [
                    { id: 'p3', priceType: 'recurring', price: { unit: 'USD', value: 5 }, priceAlteration: { type: 'discount', name: 'Loyalty' } },
                    { id: 'p4', priceType: 'usage', price: { unit: 'USD', value: 1 }, priceAlteration: { type: 'fee', name: 'Setup' } },
                ],
            },
            isLoading: false,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();

        expect(screen.getByText(/Discount: Loyalty/)).toBeInTheDocument();
        expect(screen.getByText(/Fee: Setup/)).toBeInTheDocument();
        expect(screen.getByText('Usage')).toBeInTheDocument();
        expect(screen.getByText(/Start -/)).toBeInTheDocument();
    });

    it('shows an empty state when there are no prices, and renders attachments and a spec link', () => {
        vi.mocked(api.useOffering).mockReturnValue({
            data: {
                ...sampleOffering,
                isBundle: true,
                productOfferingPrice: [],
                categories: [],
                attachments: [
                    { id: 'a1', name: 'Datasheet', url: 'https://example.com/d.pdf', type: 'Document', description: 'Spec sheet' },
                ],
                productSpecification: { id: 'spec-1', name: 'Fiber Spec', productNumber: 'SKU-1', isBundle: false, lifecycleStatus: 'Active', validFor: {}, lastUpdate: '' },
            },
            isLoading: false,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();

        expect(screen.getByText('No prices defined for this offering.')).toBeInTheDocument();
        expect(screen.getByRole('link', { name: 'Datasheet' })).toHaveAttribute('href', 'https://example.com/d.pdf');
        expect(screen.getByRole('link', { name: /view specification/i })).toHaveAttribute(
            'href',
            '/catalog/specifications/spec-1'
        );
    });
});
