import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import SpecificationDetailPage from './SpecificationDetailPage';
import * as api from './api';
import type { LifecycleStatus, ProductSpecification } from './types';

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
    };
});

const renderComponent = (initialEntries = ['/catalog/specifications/spec-1']) => {
    return render(
        <QueryClientProvider client={queryClient}>
            <MemoryRouter initialEntries={initialEntries}>
                <Routes>
                    <Route path="/catalog/specifications/:id" element={<SpecificationDetailPage />} />
                </Routes>
            </MemoryRouter>
        </QueryClientProvider>
    );
};

const sampleSpec: ProductSpecification = {
    id: 'spec-1',
    name: 'Fiber Modem',
    description: 'A fast modem',
    productNumber: 'SKU-100',
    isBundle: true,
    lifecycleStatus: 'Active',
    validFor: { startDateTime: '2026-01-01T00:00:00Z' },
    lastUpdate: '2026-06-01T00:00:00Z',
    characteristics: {
        speed: {
            name: 'Speed',
            description: 'Download speed',
            valueType: 'string',
            configurable: true,
            validValues: ['100Mbps', '1Gbps'],
        },
    },
};

describe('SpecificationDetailPage', () => {
    it('renders the spec with characteristics, bundle badge and edit link', () => {
        vi.mocked(api.useSpecification).mockReturnValue({
            data: sampleSpec,
            isLoading: false,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();

        expect(screen.getByRole('heading', { name: 'Fiber Modem' })).toBeInTheDocument();
        expect(document.querySelector('.bundle-badge')).toHaveTextContent('Bundle');
        expect(screen.getByText('Speed')).toBeInTheDocument();
        expect(screen.getByText('100Mbps')).toBeInTheDocument();
        expect(screen.getByText('Configurable')).toBeInTheDocument();
        expect(screen.getByRole('link', { name: /edit specification/i })).toHaveAttribute(
            'href',
            '/catalog/specifications/spec-1/edit'
        );
    });

    it('renders a minimal characteristic and open-ended validity with no description', () => {
        vi.mocked(api.useSpecification).mockReturnValue({
            data: {
                ...sampleSpec,
                description: undefined,
                validFor: { endDateTime: '2027-01-01T00:00:00Z' },
                characteristics: {
                    color: { name: 'Color', valueType: 'string', configurable: false },
                },
            },
            isLoading: false,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();

        expect(screen.getByText('No description provided.')).toBeInTheDocument();
        expect(screen.getByText('Color')).toBeInTheDocument();
        expect(screen.queryByText('Configurable')).not.toBeInTheDocument();
        expect(screen.getByText(/Start -/)).toBeInTheDocument();
    });

    it('shows an empty state when there are no characteristics', () => {
        vi.mocked(api.useSpecification).mockReturnValue({
            data: { ...sampleSpec, isBundle: false, characteristics: {} },
            isLoading: false,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();

        expect(screen.getByText('No characteristics defined for this specification.')).toBeInTheDocument();
    });

    it('shows a loading state while fetching', () => {
        vi.mocked(api.useSpecification).mockReturnValue({
            data: undefined,
            isLoading: true,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();

        expect(screen.getByText('Loading specification...')).toBeInTheDocument();
    });

    it('shows a not-found state when the spec is missing', () => {
        vi.mocked(api.useSpecification).mockReturnValue({
            data: undefined,
            isLoading: false,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();

        expect(screen.getByText('Specification not found')).toBeInTheDocument();
    });

    it('shows an error message when the query fails', () => {
        vi.mocked(api.useSpecification).mockReturnValue({
            data: undefined,
            isLoading: false,
            error: new Error('nope'),
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        renderComponent();

        expect(screen.getByText('Error: nope')).toBeInTheDocument();
    });

    it.each(['Active', 'Retired', 'Draft', 'Suspended'] as LifecycleStatus[])(
        'renders the %s status icon branch',
        (status) => {
            vi.mocked(api.useSpecification).mockReturnValue({
                data: { ...sampleSpec, lifecycleStatus: status },
                isLoading: false,
                error: null,
            } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

            renderComponent();

            expect(screen.getAllByText(status).length).toBeGreaterThan(0);
        }
    );
});
