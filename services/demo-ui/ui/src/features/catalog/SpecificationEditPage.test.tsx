import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import SpecificationEditPage from './SpecificationEditPage';
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
        useSpecification: vi.fn(),
        useCreateSpecification: vi.fn(() => ({ mutateAsync: vi.fn(), isPending: false })),
        useUpdateSpecification: vi.fn(() => ({ mutateAsync: vi.fn(), isPending: false })),
    };
});

const renderComponent = (initialEntries = ['/catalog/specifications/new']) => {
    return render(
        <QueryClientProvider client={queryClient}>
            <MemoryRouter initialEntries={initialEntries}>
                <Routes>
                    <Route path="/catalog/specifications/new" element={<SpecificationEditPage />} />
                    <Route path="/catalog/specifications/:id" element={<SpecificationEditPage />} />
                </Routes>
            </MemoryRouter>
        </QueryClientProvider>
    );
};

describe('SpecificationEditPage', () => {
    it('renders "New Product Specification" form when creating new spec', async () => {
        vi.mocked(api.useSpecification).mockReturnValue({
            data: undefined,
            isLoading: false,
            isError: false,
        } as any);

        renderComponent(['/catalog/specifications/new']);

        // This is expected to FAIL currently
        expect(screen.getByText('New Product Specification')).toBeInTheDocument();
        expect(screen.queryByText('Specification not found')).not.toBeInTheDocument();
    });
});
