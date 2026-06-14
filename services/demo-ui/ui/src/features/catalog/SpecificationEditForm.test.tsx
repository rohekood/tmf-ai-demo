import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import SpecificationEditForm from './SpecificationEditForm';
import * as api from './api';
import type { ProductSpecification } from './types';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
    return { ...actual, useNavigate: () => mockNavigate };
});

vi.mock('./components/CharacteristicEditor', () => ({
    default: ({ onChange }: { onChange: (c: Record<string, unknown>) => void }) => (
        <button type="button" onClick={() => onChange({ color: { name: 'Color', valueType: 'string', configurable: true } })}>
            edit-chars
        </button>
    ),
}));

vi.mock('./api', async () => {
    const actual = await vi.importActual('./api');
    return {
        ...actual,
        useCreateSpecification: vi.fn(),
        useUpdateSpecification: vi.fn(),
    };
});

const existing: ProductSpecification = {
    id: 'spec-1',
    name: 'Modem',
    description: 'd',
    productNumber: 'SKU-1',
    isBundle: false,
    lifecycleStatus: 'Active',
    validFor: { startDateTime: '2026-01-01T00:00:00Z' },
    lastUpdate: '2026-06-01T00:00:00Z',
};

function setup(createMock = vi.fn(), updateMock = vi.fn(), isPending = false) {
    vi.mocked(api.useCreateSpecification).mockReturnValue({ mutateAsync: createMock, isPending } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
    vi.mocked(api.useUpdateSpecification).mockReturnValue({ mutateAsync: updateMock, isPending: false } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
}

const renderForm = (props: { specification?: ProductSpecification; isNew: boolean }) =>
    render(
        <MemoryRouter>
            <SpecificationEditForm {...props} />
        </MemoryRouter>
    );

beforeEach(() => mockNavigate.mockClear());

describe('SpecificationEditForm', () => {
    it('creates a new specification with characteristics and bundle flag', async () => {
        const createMock = vi.fn().mockResolvedValue({ id: 'spec-new' });
        setup(createMock);
        renderForm({ isNew: true });

        expect(screen.getByText('New Product Specification')).toBeInTheDocument();
        fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Router' } });
        fireEvent.change(screen.getByLabelText('Product Number (SKU)'), { target: { value: 'SKU-9' } });
        fireEvent.change(screen.getByLabelText('Lifecycle Status'), { target: { value: 'Active' } });
        fireEvent.change(screen.getByLabelText('Description'), { target: { value: 'A description' } });
        fireEvent.change(screen.getByLabelText('End Date'), { target: { value: '2026-12-31' } });
        fireEvent.click(screen.getByLabelText('Is Bundle'));
        fireEvent.click(screen.getByText('edit-chars'));
        fireEvent.change(screen.getByLabelText('Start Date'), { target: { value: '2026-06-14' } });
        fireEvent.submit(document.getElementById('spec-form')!);

        await waitFor(() => expect(createMock).toHaveBeenCalled());
        expect(createMock).toHaveBeenCalledWith(expect.objectContaining({
            name: 'Router',
            productNumber: 'SKU-9',
            isBundle: true,
            characteristics: { color: { name: 'Color', valueType: 'string', configurable: true } },
            validFor: { startDateTime: '2026-06-14T00:00:00Z', endDateTime: '2026-12-31T23:59:59Z' },
        }));
        expect(mockNavigate).toHaveBeenCalledWith('/catalog/specifications/spec-new');
    });

    it('updates an existing specification', async () => {
        const updateMock = vi.fn().mockResolvedValue({});
        setup(vi.fn(), updateMock);
        renderForm({ specification: existing, isNew: false });

        expect(screen.getByText('Edit Specification')).toBeInTheDocument();
        expect(screen.getByText('Editing Modem')).toBeInTheDocument();
        fireEvent.submit(document.getElementById('spec-form')!);

        await waitFor(() => expect(updateMock).toHaveBeenCalledWith(expect.objectContaining({ id: 'spec-1' })));
        expect(mockNavigate).toHaveBeenCalledWith('/catalog/specifications/spec-1');
    });

    it('logs an error when saving fails', async () => {
        const createMock = vi.fn().mockRejectedValue(new Error('fail'));
        const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
        setup(createMock);
        renderForm({ isNew: true });
        fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'X' } });
        fireEvent.submit(document.getElementById('spec-form')!);
        await waitFor(() => expect(errorSpy).toHaveBeenCalledWith('Failed to save specification:', expect.any(Error)));
        errorSpy.mockRestore();
    });

    it('disables submit while pending', () => {
        setup(vi.fn(), vi.fn(), true);
        renderForm({ isNew: true });
        expect(screen.getByRole('button', { name: /create specification/i })).toBeDisabled();
    });
});
