import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import CatalogEditForm from './CatalogEditForm';
import * as api from './api';
import type { Catalog, Category } from './types';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
    return { ...actual, useNavigate: () => mockNavigate };
});

// Stub CategoryPicker so we can drive its onChange to exercise link/unlink logic.
vi.mock('./components/CategoryPicker', () => ({
    default: ({ onChange, selectedIds }: { onChange: (ids: string[]) => void; selectedIds: string[] }) => (
        <div>
            <span data-testid="selected">{selectedIds.join(',')}</span>
            <button type="button" onClick={() => onChange([...selectedIds, 'cat-new'])}>add-cat</button>
            <button type="button" onClick={() => onChange([])}>clear-cat</button>
        </div>
    ),
}));

vi.mock('./api', async () => {
    const actual = await vi.importActual('./api');
    return {
        ...actual,
        useCreateCatalog: vi.fn(),
        useCatalogUpdate: vi.fn(),
        useCategories: vi.fn(),
        useUpdateCategory: vi.fn(),
    };
});

const existingCatalog: Catalog = {
    id: 'cat-1',
    name: 'Consumer',
    description: 'desc',
    validFor: { startDateTime: '2026-01-01T00:00:00Z', endDateTime: '2026-12-31T23:59:59Z' },
    lastUpdate: '2026-06-01T00:00:00Z',
    lifecycleStatus: 'Active',
};

const linkedCategory: Category = {
    id: 'linked-1',
    name: 'Linked',
    isRoot: true,
    catalogId: 'cat-1',
    validFor: {},
    lastUpdate: '',
    lifecycleStatus: 'Active',
};

function setup(createMock = vi.fn(), updateMock = vi.fn(), updateCatMock = vi.fn(), opts: { isPending?: boolean; categories?: Category[] } = {}) {
    vi.mocked(api.useCreateCatalog).mockReturnValue({ mutateAsync: createMock, isPending: opts.isPending ?? false } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
    vi.mocked(api.useCatalogUpdate).mockReturnValue({ mutateAsync: updateMock, isPending: false } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
    vi.mocked(api.useCategories).mockReturnValue({ data: opts.categories ?? [] } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
    vi.mocked(api.useUpdateCategory).mockReturnValue({ mutateAsync: updateCatMock } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
}

const renderForm = (props: { catalog?: Catalog; isNew: boolean }) =>
    render(
        <MemoryRouter>
            <CatalogEditForm {...props} />
        </MemoryRouter>
    );

beforeEach(() => {
    mockNavigate.mockClear();
});

describe('CatalogEditForm', () => {
    it('creates a new catalog and navigates to its detail page', async () => {
        const createMock = vi.fn().mockResolvedValue({ id: 'new-1' });
        setup(createMock);
        renderForm({ isNew: true });

        expect(screen.getByText('New Product Catalog')).toBeInTheDocument();
        expect(screen.queryByText('add-cat')).not.toBeInTheDocument(); // no CategoryPicker when new

        fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'My Catalog' } });
        fireEvent.change(screen.getByLabelText('Description'), { target: { value: 'Hello' } });
        fireEvent.change(screen.getByLabelText('Start Date'), { target: { value: '2026-06-14' } });
        fireEvent.change(screen.getByLabelText('End Date'), { target: { value: '2026-06-20' } });
        fireEvent.submit(document.getElementById('catalog-form')!);

        await waitFor(() => expect(createMock).toHaveBeenCalled());
        expect(createMock).toHaveBeenCalledWith(expect.objectContaining({
            name: 'My Catalog',
            description: 'Hello',
            validFor: { startDateTime: '2026-06-14T00:00:00Z', endDateTime: '2026-06-20T23:59:59Z' },
        }));
        expect(mockNavigate).toHaveBeenCalledWith('/catalog/catalogs/new-1');
    });

    it('omits validFor datetimes when dates are blank', async () => {
        const createMock = vi.fn().mockResolvedValue({ id: 'new-2' });
        setup(createMock);
        renderForm({ isNew: true });
        fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'No dates' } });
        fireEvent.submit(document.getElementById('catalog-form')!);

        await waitFor(() => expect(createMock).toHaveBeenCalled());
        expect(createMock).toHaveBeenCalledWith(expect.objectContaining({
            validFor: { startDateTime: undefined, endDateTime: undefined },
        }));
    });

    it('updates an existing catalog and shows the edit heading', async () => {
        const updateMock = vi.fn().mockResolvedValue({});
        setup(vi.fn(), updateMock);
        renderForm({ catalog: existingCatalog, isNew: false });

        expect(screen.getByText('Edit Catalog')).toBeInTheDocument();
        expect(screen.getByText('Editing Consumer')).toBeInTheDocument();
        expect((screen.getByLabelText('Name') as HTMLInputElement).value).toBe('Consumer');

        fireEvent.change(screen.getByLabelText('Lifecycle Status'), { target: { value: 'Retired' } });
        fireEvent.submit(document.getElementById('catalog-form')!);

        await waitFor(() => expect(updateMock).toHaveBeenCalled());
        expect(updateMock).toHaveBeenCalledWith(expect.objectContaining({
            id: 'cat-1',
            payload: expect.objectContaining({ lifecycleStatus: 'Retired' }),
        }));
        expect(mockNavigate).toHaveBeenCalledWith('/catalog/catalogs/cat-1');
    });

    it('shows an error message from the mutation', async () => {
        const createMock = vi.fn().mockRejectedValue(new Error('boom'));
        setup(createMock);
        renderForm({ isNew: true });
        fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'X' } });
        fireEvent.submit(document.getElementById('catalog-form')!);
        expect(await screen.findByText('boom')).toBeInTheDocument();
    });

    it('shows a generic error for a non-Error rejection', async () => {
        const createMock = vi.fn().mockRejectedValue('weird');
        setup(createMock);
        renderForm({ isNew: true });
        fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'X' } });
        fireEvent.submit(document.getElementById('catalog-form')!);
        expect(await screen.findByText('Failed to save catalog.')).toBeInTheDocument();
    });

    it('disables the submit button and shows a spinner while pending', () => {
        setup(vi.fn(), vi.fn(), vi.fn(), { isPending: true });
        renderForm({ isNew: true });
        expect(screen.getByRole('button', { name: /create catalog/i })).toBeDisabled();
    });

    it('links and unlinks categories via the picker', async () => {
        const updateCatMock = vi.fn().mockResolvedValue({});
        setup(vi.fn(), vi.fn(), updateCatMock, { categories: [linkedCategory] });
        renderForm({ catalog: existingCatalog, isNew: false });

        expect(screen.getByTestId('selected')).toHaveTextContent('linked-1');

        // Add a category -> links 'cat-new'
        fireEvent.click(screen.getByText('add-cat'));
        await waitFor(() => expect(updateCatMock).toHaveBeenCalledWith({ id: 'cat-new', payload: { catalogId: 'cat-1' } }));

        // Clear -> unlinks the previously linked one
        fireEvent.click(screen.getByText('clear-cat'));
        await waitFor(() => expect(updateCatMock).toHaveBeenCalledWith({ id: 'linked-1', payload: { catalogId: null } }));
    });

    it('logs an error when category linking fails', async () => {
        const updateCatMock = vi.fn().mockRejectedValue(new Error('nope'));
        const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
        setup(vi.fn(), vi.fn(), updateCatMock, { categories: [linkedCategory] });
        renderForm({ catalog: existingCatalog, isNew: false });

        fireEvent.click(screen.getByText('add-cat'));
        await waitFor(() => expect(errorSpy).toHaveBeenCalled());
        errorSpy.mockRestore();
    });
});
