import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import OfferingEditPage from './OfferingEditPage';
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
        useOffering: vi.fn(),
        useCreateOffering: vi.fn(() => ({ mutateAsync: vi.fn(), isPending: false })),
        useUpdateOffering: vi.fn(() => ({ mutateAsync: vi.fn(), isPending: false })),
        useSpecifications: vi.fn(() => ({ data: [] })),
        useCategories: vi.fn(() => ({ data: [] })),
    };
});

const renderComponent = (initialEntries = ['/catalog/offerings/new']) => {
    return render(
        <QueryClientProvider client={queryClient}>
            <MemoryRouter initialEntries={initialEntries}>
                <Routes>
                    <Route path="/catalog/offerings/new" element={<OfferingEditPage />} />
                    <Route path="/catalog/offerings/:id" element={<OfferingEditPage />} />
                </Routes>
            </MemoryRouter>
        </QueryClientProvider>
    );
};

describe('OfferingEditPage', () => {
    it('renders "New Product Offering" form when creating new offering', async () => {
        vi.mocked(api.useOffering).mockReturnValue({
            data: undefined,
            isLoading: false,
            isError: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent(['/catalog/offerings/new']);

        // This is expected to FAIL currently
        // Note check Setting header text in OfferingEditForm if it matches "New Product Offering"
        // Based on other forms it's likely "New Product Offering" or similar.
        // Let's verify OfferingEditForm content or just check for "Offering not found" absence.
        // Assuming "New Product Offering" is the header.

        expect(screen.queryByText('Offering not found')).not.toBeInTheDocument();
        expect(screen.getByRole('button', { name: /create offering/i })).toBeInTheDocument();
    });
});
