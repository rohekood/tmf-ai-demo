import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import CatalogEditPage from './CatalogEditPage';
import * as api from './api';

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
        useCreateCatalog: vi.fn(() => ({ mutateAsync: vi.fn(), isPending: false })),
        useCatalogUpdate: vi.fn(() => ({ mutateAsync: vi.fn(), isPending: false })),
        useCategories: vi.fn(() => ({ data: [] })),
        useCreateCategory: vi.fn(() => ({ mutateAsync: vi.fn() })),
        useDeleteCategory: vi.fn(() => ({ mutateAsync: vi.fn() })),
        useUpdateCategory: vi.fn(() => ({ mutateAsync: vi.fn() })),
    };
});

const renderComponent = (initialEntries = ['/catalog/catalogs/new']) => {
    return render(
        <QueryClientProvider client={queryClient}>
            <MemoryRouter initialEntries={initialEntries}>
                <Routes>
                    <Route path="/catalog/catalogs/new" element={<CatalogEditPage />} />
                    <Route path="/catalog/catalogs/:id" element={<CatalogEditPage />} />
                </Routes>
            </MemoryRouter>
        </QueryClientProvider>
    );
};

describe('CatalogEditPage', () => {
    it('renders "New Product Catalog" form when creating new catalog', async () => {
        // Mock asking for "new" or undefined ID
        vi.mocked(api.useCatalog).mockReturnValue({
            data: undefined,
            isLoading: false,
            isError: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent(['/catalog/catalogs/new']);

        // This is expected to FAIL currently because of the bug
        expect(screen.getByText('New Product Catalog')).toBeInTheDocument();
        expect(screen.queryByText('Catalog not found')).not.toBeInTheDocument();
    });
});
