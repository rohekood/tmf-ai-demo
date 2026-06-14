import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import OfferingEditForm from './OfferingEditForm';
import * as api from './api';
import type { ProductOffering, ProductSpecification } from './types';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
    return { ...actual, useNavigate: () => mockNavigate };
});

vi.mock('./components/CategoryPicker', () => ({
    default: ({ onChange }: { onChange: (ids: string[]) => void }) => (
        <button type="button" onClick={() => onChange(['cat-1'])}>set-cats</button>
    ),
}));
vi.mock('./components/PriceEditor', () => ({
    default: ({ onChange }: { onChange: (p: unknown[]) => void }) => (
        <button type="button" onClick={() => onChange([{ priceType: 'one_time', price: { unit: 'EUR', value: 10 } }])}>set-prices</button>
    ),
}));
vi.mock('./components/AttachmentManager', () => ({
    default: ({ onChange }: { onChange: (a: unknown[]) => void }) => (
        <button type="button" onClick={() => onChange([{ id: 'a1', name: 'f', url: 'u', type: 'Document' }])}>set-atts</button>
    ),
}));

vi.mock('./api', async () => {
    const actual = await vi.importActual('./api');
    return {
        ...actual,
        useCreateOffering: vi.fn(),
        useUpdateOffering: vi.fn(),
        useSpecifications: vi.fn(),
    };
});

const spec: ProductSpecification = {
    id: 'spec-1', name: 'Modem', productNumber: 'SKU-1', isBundle: false,
    lifecycleStatus: 'Active', validFor: {}, lastUpdate: '',
};

const existing: ProductOffering = {
    id: 'off-1', name: 'Plan', description: 'd', lifecycleStatus: 'Active',
    validFor: { startDateTime: '2026-01-01T00:00:00Z' }, lastUpdate: '',
    isBundle: false, isSellable: true, productSpecificationId: 'spec-1',
    productOfferingPrice: [], categoryIds: [], attachments: [],
};

function setup(createMock = vi.fn(), updateMock = vi.fn(), isPending = false) {
    vi.mocked(api.useCreateOffering).mockReturnValue({ mutateAsync: createMock, isPending } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
    vi.mocked(api.useUpdateOffering).mockReturnValue({ mutateAsync: updateMock, isPending: false } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
    vi.mocked(api.useSpecifications).mockReturnValue({ data: [spec] } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
}

const renderForm = (props: { offering?: ProductOffering; isNew: boolean }) =>
    render(
        <MemoryRouter>
            <OfferingEditForm {...props} />
        </MemoryRouter>
    );

beforeEach(() => mockNavigate.mockClear());

describe('OfferingEditForm', () => {
    it('creates a new offering with spec, prices, categories and attachments', async () => {
        const createMock = vi.fn().mockResolvedValue({ id: 'off-new' });
        setup(createMock);
        renderForm({ isNew: true });

        expect(screen.getByText('New Product Offering')).toBeInTheDocument();
        fireEvent.change(screen.getByLabelText('Offering Name'), { target: { value: 'Fiber' } });
        fireEvent.change(screen.getByLabelText('Product Specification'), { target: { value: 'spec-1' } });
        fireEvent.change(screen.getByLabelText('Lifecycle Status'), { target: { value: 'Active' } });
        fireEvent.change(screen.getByLabelText('Description'), { target: { value: 'A description' } });
        fireEvent.change(screen.getByLabelText('Start Date'), { target: { value: '2026-01-01' } });
        fireEvent.click(screen.getByLabelText('Bundle'));
        fireEvent.click(screen.getByLabelText('Sellable')); // toggle off
        fireEvent.click(screen.getByText('set-cats'));
        fireEvent.click(screen.getByText('set-prices'));
        fireEvent.click(screen.getByText('set-atts'));
        fireEvent.change(screen.getByLabelText('End Date'), { target: { value: '2026-12-31' } });
        fireEvent.submit(document.getElementById('offering-form')!);

        await waitFor(() => expect(createMock).toHaveBeenCalled());
        expect(createMock).toHaveBeenCalledWith(expect.objectContaining({
            name: 'Fiber',
            productSpecificationId: 'spec-1',
            isBundle: true,
            isSellable: false,
            categoryIds: ['cat-1'],
            productOfferingPrice: [{ priceType: 'one_time', price: { unit: 'EUR', value: 10 } }],
            attachments: [{ id: 'a1', name: 'f', url: 'u', type: 'Document' }],
            validFor: { startDateTime: '2026-01-01T00:00:00Z', endDateTime: '2026-12-31T23:59:59Z' },
        }));
        expect(mockNavigate).toHaveBeenCalledWith('/catalog/offerings/off-new');
    });

    it('updates an existing offering', async () => {
        const updateMock = vi.fn().mockResolvedValue({});
        setup(vi.fn(), updateMock);
        renderForm({ offering: existing, isNew: false });

        expect(screen.getByText('Edit Offering')).toBeInTheDocument();
        expect(screen.getByText('Editing Plan')).toBeInTheDocument();
        fireEvent.submit(document.getElementById('offering-form')!);

        await waitFor(() => expect(updateMock).toHaveBeenCalledWith(expect.objectContaining({ id: 'off-1' })));
        expect(mockNavigate).toHaveBeenCalledWith('/catalog/offerings/off-1');
    });

    it('logs an error when saving fails', async () => {
        const createMock = vi.fn().mockRejectedValue(new Error('fail'));
        const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
        setup(createMock);
        renderForm({ isNew: true });
        fireEvent.change(screen.getByLabelText('Offering Name'), { target: { value: 'X' } });
        fireEvent.submit(document.getElementById('offering-form')!);
        await waitFor(() => expect(errorSpy).toHaveBeenCalledWith('Failed to save offering:', expect.any(Error)));
        errorSpy.mockRestore();
    });

    it('disables submit while pending', () => {
        setup(vi.fn(), vi.fn(), true);
        renderForm({ isNew: true });
        expect(screen.getByRole('button', { name: /create offering/i })).toBeDisabled();
    });
});
