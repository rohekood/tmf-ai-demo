import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import CategoryEditPage from './CategoryEditPage';
import * as api from './api';
import type { Category } from './types';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
    return { ...actual, useNavigate: () => mockNavigate };
});

vi.mock('./api', async () => {
    const actual = await vi.importActual('./api');
    return {
        ...actual,
        useCategory: vi.fn(),
        useCategories: vi.fn(),
        useCreateCategory: vi.fn(),
        useUpdateCategory: vi.fn(),
    };
});

const parent: Category = { id: 'root-1', name: 'Root', isRoot: true, validFor: {}, lastUpdate: '', lifecycleStatus: 'Active' };
const child: Category = { id: 'cat-9', name: 'Child', description: 'd', parentId: 'root-1', isRoot: false, validFor: {}, lastUpdate: '', lifecycleStatus: 'Active' };

function setup(opts: { category?: Category; isCategoryLoading?: boolean; createMock?: ReturnType<typeof vi.fn>; updateMock?: ReturnType<typeof vi.fn>; isPending?: boolean } = {}) {
    vi.mocked(api.useCategory).mockReturnValue({ data: opts.category, isLoading: opts.isCategoryLoading ?? false } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
    vi.mocked(api.useCategories).mockReturnValue({ data: [parent, child] } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
    vi.mocked(api.useCreateCategory).mockReturnValue({ mutateAsync: opts.createMock ?? vi.fn(), isPending: opts.isPending ?? false } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
    vi.mocked(api.useUpdateCategory).mockReturnValue({ mutateAsync: opts.updateMock ?? vi.fn(), isPending: false } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
}

const renderNew = () => render(
    <MemoryRouter initialEntries={['/catalog/categories/new']}>
        <Routes><Route path="/catalog/categories/new" element={<CategoryEditPage />} /></Routes>
    </MemoryRouter>
);
const renderEdit = (id = 'cat-9') => render(
    <MemoryRouter initialEntries={[`/catalog/categories/${id}/edit`]}>
        <Routes><Route path="/catalog/categories/:id/edit" element={<CategoryEditPage />} /></Routes>
    </MemoryRouter>
);

beforeEach(() => mockNavigate.mockClear());

describe('CategoryEditPage', () => {
    it('creates a new root category', async () => {
        const createMock = vi.fn().mockResolvedValue({ id: 'new' });
        setup({ createMock });
        renderNew();

        expect(screen.getByRole('heading', { name: 'New Category' })).toBeInTheDocument();
        fireEvent.change(screen.getByLabelText('Category Name'), { target: { value: 'Electronics' } });
        fireEvent.change(screen.getByLabelText('Description'), { target: { value: 'All electronics' } });
        fireEvent.click(screen.getByLabelText('Is Root Category?'));
        fireEvent.change(screen.getByLabelText('Start Date'), { target: { value: '2026-06-14' } });
        fireEvent.change(screen.getByLabelText('End Date'), { target: { value: '2026-12-31' } });
        fireEvent.submit(document.getElementById('category-form')!);

        await waitFor(() => expect(createMock).toHaveBeenCalled());
        expect(createMock).toHaveBeenCalledWith(expect.objectContaining({
            name: 'Electronics', isRoot: true, parentId: undefined,
            validFor: expect.objectContaining({ startDateTime: '2026-06-14T00:00:00Z' }),
        }));
        expect(mockNavigate).toHaveBeenCalledWith('/catalog/categories');
    });

    it('creates a child category with a selected parent', async () => {
        const createMock = vi.fn().mockResolvedValue({ id: 'new' });
        setup({ createMock });
        renderNew();
        fireEvent.change(screen.getByLabelText('Category Name'), { target: { value: 'Phones' } });
        fireEvent.change(screen.getByLabelText('Parent Category'), { target: { value: 'root-1' } });
        fireEvent.submit(document.getElementById('category-form')!);

        await waitFor(() => expect(createMock).toHaveBeenCalledWith(expect.objectContaining({ isRoot: false, parentId: 'root-1' })));
    });

    it('edits an existing category', async () => {
        const updateMock = vi.fn().mockResolvedValue({});
        setup({ category: child, updateMock });
        renderEdit();

        expect(screen.getByRole('heading', { name: 'Edit Category' })).toBeInTheDocument();
        expect((screen.getByLabelText('Category Name') as HTMLInputElement).value).toBe('Child');
        fireEvent.change(screen.getByLabelText('Lifecycle Status'), { target: { value: 'Retired' } });
        fireEvent.submit(document.getElementById('category-form')!);

        await waitFor(() => expect(updateMock).toHaveBeenCalledWith(expect.objectContaining({
            id: 'cat-9', payload: expect.objectContaining({ lifecycleStatus: 'Retired' }),
        })));
        expect(mockNavigate).toHaveBeenCalledWith('/catalog/categories');
    });

    it('shows the error message from a failed save', async () => {
        const createMock = vi.fn().mockRejectedValue(new Error('boom'));
        setup({ createMock });
        renderNew();
        fireEvent.change(screen.getByLabelText('Category Name'), { target: { value: 'X' } });
        fireEvent.submit(document.getElementById('category-form')!);
        expect(await screen.findByText('boom')).toBeInTheDocument();
    });

    it('shows a generic error for a non-Error rejection', async () => {
        const createMock = vi.fn().mockRejectedValue('weird');
        setup({ createMock });
        renderNew();
        fireEvent.change(screen.getByLabelText('Category Name'), { target: { value: 'X' } });
        fireEvent.submit(document.getElementById('category-form')!);
        expect(await screen.findByText('Failed to save category.')).toBeInTheDocument();
    });

    it('shows a loading state in edit mode', () => {
        setup({ isCategoryLoading: true });
        const { container } = renderEdit();
        expect(container.querySelector('.spin')).toBeInTheDocument();
    });
});
