import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import OfferingDetailPage from './OfferingDetailPage';
import * as api from './api';
import type { ProductOffering } from './types';

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

    it('shows a not-found state when the offering is missing', () => {
        vi.mocked(api.useOffering).mockReturnValue({
            data: undefined,
            isLoading: false,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();

        expect(screen.getByText('Offering not found')).toBeInTheDocument();
    });
});
