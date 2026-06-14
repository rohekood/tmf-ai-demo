import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { BrowserRouter } from 'react-router-dom';
import CatalogListPage from './CatalogListPage';
import * as api from './api';
import type { Catalog } from './types';

vi.mock('./api', async () => {
    const actual = await vi.importActual('./api');
    return { ...actual, useCatalogs: vi.fn(), useCatalogDelete: vi.fn() };
});

const catalogs: Catalog[] = [
    { id: 'c1', name: 'Consumer', description: 'home', validFor: {}, lastUpdate: new Date(2026, 5, 14).toISOString(), lifecycleStatus: 'Active' },
    { id: 'c2', name: 'Business', description: 'biz', validFor: {}, lastUpdate: new Date(2026, 0, 2).toISOString(), lifecycleStatus: 'Draft' },
];

function setup(data: Catalog[] | undefined, opts: { isLoading?: boolean; error?: Error | null; mutate?: ReturnType<typeof vi.fn>; isPending?: boolean } = {}) {
    vi.mocked(api.useCatalogs).mockReturnValue({ data, isLoading: opts.isLoading ?? false, error: opts.error ?? null } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
    vi.mocked(api.useCatalogDelete).mockReturnValue({ mutate: opts.mutate ?? vi.fn(), isPending: opts.isPending ?? false } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
}

const renderPage = () => render(<BrowserRouter><CatalogListPage /></BrowserRouter>);

beforeEach(() => vi.clearAllMocks());

describe('CatalogListPage', () => {
    it('shows a loading state', () => {
        setup(undefined, { isLoading: true });
        renderPage();
        expect(screen.getByText('Loading catalogs...')).toBeInTheDocument();
    });

    it('shows an error state', () => {
        setup(undefined, { error: new Error('down') });
        renderPage();
        expect(screen.getByText(/Failed to load catalogs: down/)).toBeInTheDocument();
    });

    it('shows an empty state with a create CTA', () => {
        setup([]);
        renderPage();
        expect(screen.getByText('No product catalogs found.')).toBeInTheDocument();
        expect(screen.getByRole('link', { name: /create your first catalog/i })).toBeInTheDocument();
    });

    it('renders rows with Estonian-formatted dates and View/Edit links', () => {
        setup(catalogs);
        renderPage();
        expect(screen.getByText('Consumer')).toBeInTheDocument();
        expect(screen.getByText('14.06.2026')).toBeInTheDocument();
        expect(screen.getAllByTitle('View')[0]).toHaveAttribute('href', '/catalog/catalogs/c1');
        expect(screen.getAllByTitle('Edit')[0]).toHaveAttribute('href', '/catalog/catalogs/c1/edit');
    });

    it('filters by search term and shows a no-match empty state', () => {
        setup(catalogs);
        renderPage();
        fireEvent.change(screen.getByPlaceholderText('Search by name or description...'), { target: { value: 'business' } });
        expect(screen.getByText('Business')).toBeInTheDocument();
        expect(screen.queryByText('Consumer')).not.toBeInTheDocument();

        fireEvent.change(screen.getByPlaceholderText('Search by name or description...'), { target: { value: 'zzz' } });
        expect(screen.getByText('No catalogs match your search.')).toBeInTheDocument();
    });

    it('sorts when a column header is clicked', () => {
        setup(catalogs);
        renderPage();
        fireEvent.click(screen.getByText('Name'));
        expect(screen.getByText('Consumer')).toBeInTheDocument();
    });

    it('deletes a catalog after confirmation', () => {
        const mutate = vi.fn();
        setup(catalogs, { mutate });
        const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true);
        renderPage();
        fireEvent.click(screen.getAllByTitle('Delete')[0]);
        expect(mutate).toHaveBeenCalledWith('c1');
        confirmSpy.mockRestore();
    });

    it('does not delete when confirmation is cancelled', () => {
        const mutate = vi.fn();
        setup(catalogs, { mutate });
        const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(false);
        renderPage();
        fireEvent.click(screen.getAllByTitle('Delete')[0]);
        expect(mutate).not.toHaveBeenCalled();
        confirmSpy.mockRestore();
    });
});
