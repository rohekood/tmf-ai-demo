import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import CustomerEditForm from './CustomerEditForm';
import * as api from './api';
import type { Customer } from './types';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
    return { ...actual, useNavigate: () => mockNavigate };
});

vi.mock('../parties/components/PartyPicker', () => ({
    default: ({ onChange, customActions }: { onChange: (p: unknown) => void; customActions?: React.ReactNode }) => (
        <div>
            <button type="button" onClick={() => onChange({ id: 'p2', '@type': 'Organization', tradingName: 'NewOrg' })}>pick-party</button>
            {customActions}
        </div>
    ),
}));

vi.mock('./components/RelatedPartiesForm', () => ({
    default: ({ onChange }: { onChange: (items: unknown[]) => void }) => (
        <button type="button" onClick={() => onChange([{ role: 'guardian', partyId: 'rp1' }])}>set-related</button>
    ),
}));

vi.mock('./api', async () => {
    const actual = await vi.importActual('./api');
    return { ...actual, useUpdateCustomer: vi.fn() };
});

const customer: Customer = {
    id: 'cust-1',
    name: 'Acme',
    status: 'active',
    partyId: 'party-1',
    partyName: 'Acme Org',
    partyType: 'Organization',
};

function setup(updateMock = vi.fn(), isPending = false) {
    vi.mocked(api.useUpdateCustomer).mockReturnValue({ mutateAsync: updateMock, isPending } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
}

const renderForm = (c: Customer = customer) =>
    render(
        <MemoryRouter>
            <CustomerEditForm customer={c} />
        </MemoryRouter>
    );

beforeEach(() => mockNavigate.mockClear());

describe('CustomerEditForm', () => {
    it('saves with no party change (partyId omitted) and empty credit profiles', async () => {
        const updateMock = vi.fn().mockResolvedValue({});
        setup(updateMock);
        renderForm();

        expect(screen.getByText('Edit Customer')).toBeInTheDocument();
        fireEvent.change(screen.getByLabelText('Customer Name'), { target: { value: 'Acme 2' } });
        fireEvent.change(screen.getByLabelText('Status'), { target: { value: 'suspended' } });
        fireEvent.submit(document.getElementById('edit-form')!);

        await waitFor(() => expect(updateMock).toHaveBeenCalled());
        expect(updateMock).toHaveBeenCalledWith(expect.objectContaining({
            id: 'cust-1', name: 'Acme 2', status: 'suspended', partyId: undefined, creditProfiles: [],
        }));
        expect(mockNavigate).toHaveBeenCalledWith('/customers/cust-1');
    });

    it('changes and reverts the linked party', async () => {
        const updateMock = vi.fn().mockResolvedValue({});
        setup(updateMock);
        renderForm();

        fireEvent.click(screen.getByText('pick-party'));
        // Revert action now visible (customActions rendered because partyId changed)
        const revert = await screen.findByText('Revert');
        fireEvent.click(revert);
        expect(screen.queryByText('Revert')).not.toBeInTheDocument();
    });

    it('submits a changed party id', async () => {
        const updateMock = vi.fn().mockResolvedValue({});
        setup(updateMock);
        renderForm();
        fireEvent.click(screen.getByText('pick-party'));
        fireEvent.submit(document.getElementById('edit-form')!);
        await waitFor(() => expect(updateMock).toHaveBeenCalledWith(expect.objectContaining({ partyId: 'p2' })));
    });

    it('adds, edits and removes a credit profile', async () => {
        const updateMock = vi.fn().mockResolvedValue({});
        setup(updateMock);
        renderForm();

        fireEvent.click(screen.getByRole('button', { name: /add profile/i }));
        fireEvent.change(screen.getByLabelText('Credit Risk Score'), { target: { value: '50' } });
        fireEvent.change(screen.getByLabelText('Credit Score'), { target: { value: '700' } });
        fireEvent.change(screen.getByLabelText('Valid From'), { target: { value: '2026-06-14' } });

        fireEvent.submit(document.getElementById('edit-form')!);
        await waitFor(() => expect(updateMock).toHaveBeenCalledWith(expect.objectContaining({
            creditProfiles: [expect.objectContaining({ creditRiskScore: 50, creditScore: 700, validForStart: '2026-06-14' })],
        })));

        // Remove the profile again
        fireEvent.click(screen.getByRole('button', { name: /remove profile/i }));
        expect(screen.queryByLabelText('Credit Score')).not.toBeInTheDocument();
    });

    it('adds, edits and removes accounts', () => {
        setup();
        renderForm();
        expect(screen.getByText('No accounts added')).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: /add account/i }));
        fireEvent.change(screen.getByLabelText('Account Name'), { target: { value: 'Main' } });
        fireEvent.change(screen.getByLabelText('Type'), { target: { value: 'Savings' } });
        fireEvent.change(document.getElementById('acc-status-0')!, { target: { value: 'suspended' } });

        // Remove (the account row delete is the last danger icon button)
        const deleteButtons = screen.getAllByRole('button').filter(b => b.className.includes('btn-icon--danger'));
        fireEvent.click(deleteButtons[deleteButtons.length - 1]);
        expect(screen.getByText('No accounts added')).toBeInTheDocument();
    });

    it('adds, edits and removes privacy consents', () => {
        setup();
        renderForm();
        expect(screen.getByText('No privacy consents')).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: /add consent/i }));
        fireEvent.change(screen.getByPlaceholderText('e.g., Marketing, Analytics'), { target: { value: 'Marketing' } });
        const statusSelect = screen.getAllByRole('combobox').find(s => (s as HTMLSelectElement).value === 'pending');
        fireEvent.change(statusSelect!, { target: { value: 'given' } });

        const deleteButtons = screen.getAllByRole('button').filter(b => b.className.includes('btn-icon--danger'));
        fireEvent.click(deleteButtons[deleteButtons.length - 1]);
        expect(screen.getByText('No privacy consents')).toBeInTheDocument();
    });

    it('updates related parties and logs an error on failed save', async () => {
        const updateMock = vi.fn().mockRejectedValue(new Error('boom'));
        const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
        setup(updateMock);
        renderForm();

        fireEvent.click(screen.getByText('set-related'));
        fireEvent.submit(document.getElementById('edit-form')!);
        await waitFor(() => expect(errorSpy).toHaveBeenCalledWith('Failed to update customer:', expect.any(Error)));
        errorSpy.mockRestore();
    });

    it('disables save while pending', () => {
        setup(vi.fn(), true);
        renderForm();
        expect(screen.getByRole('button', { name: /save changes/i })).toBeDisabled();
    });

    it('covers fallback branches (empty party, missing consent status, non-numeric credit scores)', () => {
        setup();
        renderForm({
            ...customer,
            partyId: '',
            privacyConsents: [{ consentType: 'NoStatus' }] as unknown as Customer['privacyConsents'],
        });

        // credit profile with non-numeric input -> parseInt(...) || 0
        fireEvent.click(screen.getByRole('button', { name: /add profile/i }));
        fireEvent.change(screen.getByLabelText('Credit Risk Score'), { target: { value: 'abc' } });
        fireEvent.change(screen.getByLabelText('Credit Score'), { target: { value: 'xyz' } });
        expect((screen.getByLabelText('Credit Risk Score') as HTMLInputElement).value).toBe('0');
        expect((screen.getByLabelText('Credit Score') as HTMLInputElement).value).toBe('0');
        // consent without a status falls back to 'pending'
        expect((screen.getByDisplayValue('Pending') as HTMLSelectElement)).toBeInTheDocument();
    });

    it('pre-populates existing accounts, consents and credit profile', () => {
        setup();
        renderForm({
            ...customer,
            accounts: [{ id: 'ac1', name: 'Acc1', accountType: 'Savings', accountStatus: 'active' }],
            privacyConsents: [{ id: 'pc1', consentType: 'Marketing', status: 'given' }],
            relatedParties: [{ id: 'rp1', relatedPartyId: 'rp', role: 'guardian', name: 'G' }],
            creditProfiles: [{ id: 'cp1', creditRiskScore: 42, creditScore: 800, validForStart: '2026-01-01T00:00:00Z' }],
        });

        expect((screen.getByLabelText('Account Name') as HTMLInputElement).value).toBe('Acc1');
        expect((screen.getByLabelText('Credit Risk Score') as HTMLInputElement).value).toBe('42');
        expect((screen.getByLabelText('Credit Score') as HTMLInputElement).value).toBe('800');
        expect((screen.getByLabelText('Valid From') as HTMLInputElement).value).toBe('2026-01-01');
        expect((screen.getByDisplayValue('Marketing') as HTMLInputElement)).toBeInTheDocument();
    });
});
