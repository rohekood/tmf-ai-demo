import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import type { LifecycleStatus } from './types';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import CategoryDetailPage from './CategoryDetailPage';
import * as api from './api';
import type { Category } from './types';

const queryClient = new QueryClient({
    defaultOptions: {
        queries: { retry: false },
    },
});

vi.mock('./api', async () => {
    const actual = await vi.importActual('./api');
    return {
        ...actual,
        useCategory: vi.fn(),
        useCategories: vi.fn(),
    };
});

const renderComponent = (initialEntries = ['/catalog/categories/sub-1']) => {
    return render(
        <QueryClientProvider client={queryClient}>
            <MemoryRouter initialEntries={initialEntries}>
                <Routes>
                    <Route path="/catalog/categories/:id" element={<CategoryDetailPage />} />
                </Routes>
            </MemoryRouter>
        </QueryClientProvider>
    );
};

const parentCategory: Category = {
    id: 'root-1',
    name: 'Internet',
    isRoot: true,
    validFor: {},
    lastUpdate: '2026-06-01T00:00:00Z',
    lifecycleStatus: 'Active',
};

const subCategory: Category = {
    id: 'sub-1',
    name: 'Fiber',
    description: 'Fiber broadband',
    parentId: 'root-1',
    isRoot: false,
    validFor: { startDateTime: '2026-01-01T00:00:00Z' },
    lastUpdate: '2026-06-01T00:00:00Z',
    lifecycleStatus: 'Active',
};

describe('CategoryDetailPage', () => {
    it('renders the category display view with name, description and a parent link', () => {
        vi.mocked(api.useCategory).mockReturnValue({
            data: subCategory,
            isLoading: false,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
        vi.mocked(api.useCategories).mockReturnValue({
            data: [parentCategory, subCategory],
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();

        expect(screen.getByRole('heading', { name: 'Fiber' })).toBeInTheDocument();
        expect(screen.getByText('Fiber broadband')).toBeInTheDocument();
        expect(screen.getByRole('link', { name: /edit category/i })).toHaveAttribute(
            'href',
            '/catalog/categories/sub-1/edit'
        );
        expect(screen.getByRole('link', { name: /view parent/i })).toHaveAttribute(
            'href',
            '/catalog/categories/root-1'
        );
    });

    it('renders a root category without a parent card', () => {
        vi.mocked(api.useCategory).mockReturnValue({
            data: parentCategory,
            isLoading: false,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
        vi.mocked(api.useCategories).mockReturnValue({
            data: [parentCategory],
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent(['/catalog/categories/root-1']);

        expect(screen.getByText('Root')).toBeInTheDocument();
        expect(screen.queryByRole('link', { name: /view parent/i })).not.toBeInTheDocument();
    });

    it('shows a loading state while fetching', () => {
        vi.mocked(api.useCategory).mockReturnValue({
            data: undefined,
            isLoading: true,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
        vi.mocked(api.useCategories).mockReturnValue({ data: [] } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();

        expect(screen.getByText('Loading category...')).toBeInTheDocument();
    });

    it('shows a not-found state when the category is missing', () => {
        vi.mocked(api.useCategory).mockReturnValue({
            data: undefined,
            isLoading: false,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
        vi.mocked(api.useCategories).mockReturnValue({ data: [] } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();

        expect(screen.getByText('Category not found')).toBeInTheDocument();
    });

    it('shows an error message when the query fails', () => {
        vi.mocked(api.useCategory).mockReturnValue({
            data: undefined,
            isLoading: false,
            error: new Error('kaboom'),
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
        vi.mocked(api.useCategories).mockReturnValue({ data: [] } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();

        expect(screen.getByText('Error: kaboom')).toBeInTheDocument();
    });

    it.each(['Active', 'Retired', 'Draft', 'Suspended'] as LifecycleStatus[])(
        'renders the %s status icon branch',
        (status) => {
            vi.mocked(api.useCategory).mockReturnValue({
                data: { ...parentCategory, lifecycleStatus: status },
                isLoading: false,
                error: null,
            } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
            vi.mocked(api.useCategories).mockReturnValue({ data: [parentCategory] } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

            renderComponent(['/catalog/categories/root-1']);

            expect(screen.getAllByText(status).length).toBeGreaterThan(0);
        }
    );
});
