import { render, screen, waitFor } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { BrowserRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import CategoryListPage from './CategoryListPage';
import * as api from './api';

const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            retry: false,
        },
    },
});

vi.mock('./api', async () => {
    const actual = await vi.importActual('./api');
    return {
        ...actual,
        useCategories: vi.fn(),
        useDeleteCategory: vi.fn(() => ({ mutateAsync: vi.fn() })),
    };
});

const renderComponent = () => {
    return render(
        <QueryClientProvider client={queryClient}>
            <BrowserRouter>
                <CategoryListPage />
            </BrowserRouter>
        </QueryClientProvider>
    );
};

describe('CategoryListPage', () => {
    it('shows loading state initially', () => {
        vi.mocked(api.useCategories).mockReturnValue({
            data: undefined,
            isLoading: true,
            isError: false,
        } as any);

        renderComponent();
        expect(screen.getByText('Loading categories...')).toBeInTheDocument();
    });

    it('renders categories in list', async () => {
        vi.mocked(api.useCategories).mockReturnValue({
            data: [
                { id: '1', name: 'Electronics', isRoot: true, parentId: undefined },
                { id: '2', name: 'Laptops', isRoot: false, parentId: '1' }
            ],
            isLoading: false,
            isError: false,
        } as any);

        renderComponent();
        expect(screen.getByText('Electronics')).toBeInTheDocument();

        // Expand Electronics to see Laptops
        const expandButton = screen.getByLabelText('Expand Electronics');
        expandButton.click();

        await waitFor(() => {
            expect(screen.getByText('Laptops')).toBeInTheDocument();
        });
        expect(screen.getByText('ROOT')).toBeInTheDocument(); // Badge
    });

    it('shows empty state when no categories', () => {
        vi.mocked(api.useCategories).mockReturnValue({
            data: [],
            isLoading: false,
            isError: false,
        } as any);

        renderComponent();
        expect(screen.getByText('No categories found.')).toBeInTheDocument();
    });
});
