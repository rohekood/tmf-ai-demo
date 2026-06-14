import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import type { LifecycleStatus } from './types';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import CatalogDetailPage from './CatalogDetailPage';
import * as api from './api';
import type { Catalog } from './types';

const queryClient = new QueryClient({
    defaultOptions: {
        queries: { retry: false },
    },
});

vi.mock('./api', async () => {
    const actual = await vi.importActual('./api');
    return {
        ...actual,
        useCatalog: vi.fn(),
    };
});

const renderComponent = (initialEntries = ['/catalog/catalogs/cat-1']) => {
    return render(
        <QueryClientProvider client={queryClient}>
            <MemoryRouter initialEntries={initialEntries}>
                <Routes>
                    <Route path="/catalog/catalogs/:id" element={<CatalogDetailPage />} />
                </Routes>
            </MemoryRouter>
        </QueryClientProvider>
    );
};

const sampleCatalog: Catalog = {
    id: 'cat-1',
    name: 'Consumer Catalog',
    description: 'Catalog for consumer products',
    validFor: { startDateTime: '2026-01-01T00:00:00Z' },
    lastUpdate: '2026-06-01T00:00:00Z',
    lifecycleStatus: 'Active',
};

describe('CatalogDetailPage', () => {
    it('renders the catalog display view with name, description and status', () => {
        vi.mocked(api.useCatalog).mockReturnValue({
            data: sampleCatalog,
            isLoading: false,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();

        expect(screen.getByRole('heading', { name: 'Consumer Catalog' })).toBeInTheDocument();
        expect(screen.getByText('Catalog for consumer products')).toBeInTheDocument();
        expect(screen.getAllByText('Active').length).toBeGreaterThan(0);
        expect(screen.getByRole('link', { name: /edit catalog/i })).toHaveAttribute(
            'href',
            '/catalog/catalogs/cat-1/edit'
        );
    });

    it('shows a loading state while fetching', () => {
        vi.mocked(api.useCatalog).mockReturnValue({
            data: undefined,
            isLoading: true,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();

        expect(screen.getByText('Loading catalog...')).toBeInTheDocument();
    });

    it('shows a not-found state when the catalog is missing', () => {
        vi.mocked(api.useCatalog).mockReturnValue({
            data: undefined,
            isLoading: false,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();

        expect(screen.getByText('Catalog not found')).toBeInTheDocument();
    });

    it('shows an error message when the query fails', () => {
        vi.mocked(api.useCatalog).mockReturnValue({
            data: undefined,
            isLoading: false,
            error: new Error('boom'),
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();

        expect(screen.getByText('Error: boom')).toBeInTheDocument();
    });

    it.each(['Active', 'Retired', 'Draft', 'Suspended'] as LifecycleStatus[])(
        'renders the %s status icon branch',
        (status) => {
            vi.mocked(api.useCatalog).mockReturnValue({
                data: { ...sampleCatalog, lifecycleStatus: status },
                isLoading: false,
                error: null,
            } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

            renderComponent();

            expect(screen.getAllByText(status).length).toBeGreaterThan(0);
        }
    );

    it('falls back gracefully with no description and an open-ended validity', () => {
        vi.mocked(api.useCatalog).mockReturnValue({
            data: { ...sampleCatalog, description: undefined, validFor: { endDateTime: '2027-01-01T00:00:00Z' } },
            isLoading: false,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();

        expect(screen.getByText('No description provided.')).toBeInTheDocument();
        expect(screen.getByText(/Start -/)).toBeInTheDocument();
    });
});
