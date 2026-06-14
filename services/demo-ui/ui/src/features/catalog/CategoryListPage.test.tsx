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
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

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
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

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

    it('renders View and Edit links for each category', () => {
        vi.mocked(api.useCategories).mockReturnValue({
            data: [
                { id: '1', name: 'Electronics', isRoot: true, parentId: undefined },
            ],
            isLoading: false,
            isError: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();

        expect(screen.getByTitle('View')).toHaveAttribute('href', '/catalog/categories/1');
        expect(screen.getByTitle('Edit')).toHaveAttribute('href', '/catalog/categories/1/edit');
    });

    it('shows empty state when no categories', () => {
        vi.mocked(api.useCategories).mockReturnValue({
            data: [],
            isLoading: false,
            isError: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();
        expect(screen.getByText('No categories found.')).toBeInTheDocument();
    });

    it('shows an error state when the query fails', () => {
        vi.mocked(api.useCategories).mockReturnValue({
            data: undefined,
            isLoading: false,
            isError: true,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();
        expect(screen.getByText('Error loading categories.')).toBeInTheDocument();
    });

    it('navigates to the new category page from the empty state', () => {
        vi.mocked(api.useCategories).mockReturnValue({
            data: [],
            isLoading: false,
            isError: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();
        screen.getByText('Create your first category').click();
        expect(window.location.pathname).toBe('/catalog/categories/new');
    });

    it('deletes a category after confirmation', async () => {
        const mutateAsync = vi.fn().mockResolvedValue(undefined);
        vi.mocked(api.useDeleteCategory).mockReturnValue({ mutateAsync } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
        vi.mocked(api.useCategories).mockReturnValue({
            data: [{ id: '1', name: 'Electronics', isRoot: true, parentId: undefined }],
            isLoading: false,
            isError: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
        const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true);

        renderComponent();
        screen.getByTitle('Delete category').click();

        await waitFor(() => expect(mutateAsync).toHaveBeenCalledWith('1'));
        confirmSpy.mockRestore();
    });

    it('does not delete when confirmation is cancelled', () => {
        const mutateAsync = vi.fn();
        vi.mocked(api.useDeleteCategory).mockReturnValue({ mutateAsync } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
        vi.mocked(api.useCategories).mockReturnValue({
            data: [{ id: '1', name: 'Electronics', isRoot: true, parentId: undefined }],
            isLoading: false,
            isError: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
        const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(false);

        renderComponent();
        screen.getByTitle('Delete category').click();

        expect(mutateAsync).not.toHaveBeenCalled();
        confirmSpy.mockRestore();
    });
});
