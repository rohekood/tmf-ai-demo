import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import PartyFormPage from './PartyFormPage';
import * as api from './api';
import type { Organization } from './types';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
    return { ...actual, useNavigate: () => mockNavigate };
});

// Stub PartyPicker so related-party selection can be driven.
vi.mock('./components/PartyPicker', () => ({
    default: ({ onChange }: { onChange: (p: unknown) => void }) => (
        <div>
            <button type="button" onClick={() => onChange({ id: 'rel-1', '@type': 'Organization', tradingName: 'RelOrg' })}>rp-pick</button>
            <button type="button" onClick={() => onChange(null)}>rp-clear</button>
        </div>
    ),
}));

vi.mock('./api', async () => {
    const actual = await vi.importActual('./api');
    return {
        ...actual,
        useParty: vi.fn(),
        useCreateParty: vi.fn(),
        useUpdateParty: vi.fn(),
    };
});

function setup(createMock = vi.fn(), updateMock = vi.fn(), party: unknown = undefined, isPending = false) {
    vi.mocked(api.useParty).mockReturnValue({ data: party, isLoading: false } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
    vi.mocked(api.useCreateParty).mockReturnValue({ mutateAsync: createMock, isPending } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
    vi.mocked(api.useUpdateParty).mockReturnValue({ mutateAsync: updateMock, isPending: false } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
}

const renderNew = () =>
    render(
        <MemoryRouter initialEntries={['/parties/new']}>
            <Routes>
                <Route path="/parties/new" element={<PartyFormPage />} />
            </Routes>
        </MemoryRouter>
    );

const renderEdit = (id = 'p1') =>
    render(
        <MemoryRouter initialEntries={[`/parties/${id}/edit`]}>
            <Routes>
                <Route path="/parties/:id/edit" element={<PartyFormPage />} />
            </Routes>
        </MemoryRouter>
    );

const submit = () => fireEvent.submit(document.getElementById('party-form')!);

beforeEach(() => mockNavigate.mockClear());

describe('PartyFormPage repeatable sections', () => {
    it('creates an Individual with every sub-resource populated', async () => {
        const createMock = vi.fn().mockResolvedValue({});
        setup(createMock);
        renderNew();

        fireEvent.change(screen.getByLabelText('Given Name *'), { target: { value: 'John' } });
        fireEvent.change(screen.getByLabelText('Family Name *'), { target: { value: 'Doe' } });
        fireEvent.change(screen.getByLabelText('Gender'), { target: { value: 'male' } });

        // Contact medium
        fireEvent.click(screen.getByRole('button', { name: /add contact/i }));
        fireEvent.change(screen.getByDisplayValue('Email'), { target: { value: 'phone' } });
        fireEvent.change(screen.getByPlaceholderText('+1234567890'), { target: { value: '+372 1234' } });

        // Identification
        fireEvent.click(screen.getByRole('button', { name: /add id/i }));
        fireEvent.change(screen.getByPlaceholderText('e.g., Passport, SSN'), { target: { value: 'Passport' } });
        fireEvent.change(screen.getByPlaceholderText('ID number'), { target: { value: 'X123' } });

        // External reference
        fireEvent.click(screen.getByRole('button', { name: /add reference/i }));
        fireEvent.change(screen.getByPlaceholderText('e.g. LegacyCRM'), { target: { value: 'CRM' } });
        fireEvent.change(screen.getByPlaceholderText('e.g. CUST-123'), { target: { value: 'C-1' } });

        // Attachment
        fireEvent.click(screen.getByRole('button', { name: /add attachment/i }));
        fireEvent.change(screen.getByPlaceholderText('Document Name'), { target: { value: 'Doc' } });
        fireEvent.change(screen.getByPlaceholderText('https://...'), { target: { value: 'http://x' } });
        fireEvent.change(screen.getByPlaceholderText('application/pdf'), { target: { value: 'application/pdf' } });

        // Related party
        fireEvent.click(screen.getByRole('button', { name: /add relationship/i }));
        fireEvent.click(screen.getByText('rp-pick'));
        fireEvent.change(screen.getByPlaceholderText('e.g. Employee, Next of Kin'), { target: { value: 'Employee' } });
        fireEvent.change(screen.getByPlaceholderText('read, write'), { target: { value: 'read, write' } });

        submit();

        await waitFor(() => expect(createMock).toHaveBeenCalled());
        expect(createMock).toHaveBeenCalledWith(expect.objectContaining({
            '@type': 'Individual',
            givenName: 'John',
            gender: 'male',
            contactMediums: [expect.objectContaining({ mediumType: 'phone', value: '+372 1234' })],
            identifications: [expect.objectContaining({ identificationType: 'Passport', identificationId: 'X123' })],
            externalReferences: [expect.objectContaining({ externalSystemId: 'CRM', externalReference: 'C-1' })],
            attachments: [expect.objectContaining({ name: 'Doc', url: 'http://x', mimeType: 'application/pdf' })],
            relatedParties: [expect.objectContaining({ relatedPartyId: 'rel-1', role: 'Employee', permissions: ['read', 'write'] })],
        }));
        expect(mockNavigate).toHaveBeenCalledWith('/parties');
    });

    it('clears a related party selection', () => {
        setup();
        renderNew();
        fireEvent.click(screen.getByRole('button', { name: /add relationship/i }));
        fireEvent.click(screen.getByText('rp-pick'));
        fireEvent.click(screen.getByText('rp-clear'));
        // role input still present (item not removed)
        expect(screen.getByPlaceholderText('e.g. Employee, Next of Kin')).toBeInTheDocument();
    });

    it('removes an item from every section', () => {
        setup();
        renderNew();
        const adds = [/add contact/i, /add id/i, /add exemption/i, /add reference/i, /add attachment/i, /add relationship/i];
        adds.forEach((re) => fireEvent.click(screen.getByRole('button', { name: re })));

        // Every section now has a delete (danger icon) button; remove them all.
        let danger = screen.getAllByRole('button').filter((b) => b.className.includes('btn-icon--danger'));
        expect(danger.length).toBe(6);
        while (danger.length > 0) {
            fireEvent.click(danger[0]);
            danger = screen.getAllByRole('button').filter((b) => b.className.includes('btn-icon--danger'));
        }

        expect(screen.getByText('No contact mediums added')).toBeInTheDocument();
        expect(screen.getByText('No identifications added')).toBeInTheDocument();
        expect(screen.getByText('No tax exemptions added')).toBeInTheDocument();
        expect(screen.getByText('No external references added')).toBeInTheDocument();
        expect(screen.getByText('No attachments added')).toBeInTheDocument();
        expect(screen.getByText('No related parties added')).toBeInTheDocument();
    });

    it('creates an Organization with a tax exemption and edited start date', async () => {
        const createMock = vi.fn().mockResolvedValue({});
        setup(createMock);
        renderNew();

        fireEvent.click(screen.getByRole('radio', { name: /Organization/i }));
        fireEvent.change(screen.getByLabelText('Trading Name *'), { target: { value: 'Acme' } });
        fireEvent.change(screen.getByPlaceholderText('e.g., LLC, Inc, GmbH'), { target: { value: 'LLC' } });
        fireEvent.click(screen.getByLabelText('Legal Entity')); // toggle off

        fireEvent.click(screen.getByRole('button', { name: /add exemption/i }));
        fireEvent.change(screen.getByPlaceholderText('Certificate #'), { target: { value: 'CERT-1' } });
        fireEvent.change(screen.getByPlaceholderText('Issuing Jurisdiction'), { target: { value: 'Estonia' } });
        // The only date input present in Organization mode is the tax exemption start date.
        fireEvent.change(document.querySelector('input[type="date"]')!, { target: { value: '2026-06-14' } });

        submit();

        await waitFor(() => expect(createMock).toHaveBeenCalled());
        const payload = createMock.mock.calls[0][0];
        expect(payload['@type']).toBe('Organization');
        expect(payload.tradingName).toBe('Acme');
        expect(payload.organizationType).toBe('LLC');
        expect(payload.isLegalEntity).toBe(false);
        expect(payload.taxExemptions[0].certificateNumber).toBe('CERT-1');
        expect(payload.taxExemptions[0].validFor.startDateTime).toContain('2026-06-1');
    });

    it('updates an existing Organization (edit mode) and logs an error on failure', async () => {
        const updateMock = vi.fn().mockRejectedValue(new Error('boom'));
        const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
        const org: Organization = {
            id: 'org-1',
            '@type': 'Organization',
            tradingName: 'Globex',
            isLegalEntity: true,
            organizationType: 'Inc',
            status: 'Active',
            contactMediums: [{ id: 'c1', mediumType: 'email', preferred: true, value: 'a@b.c' }],
            identifications: [],
        };
        setup(vi.fn(), updateMock, org);
        renderEdit('org-1');

        expect(screen.getByDisplayValue('Globex')).toBeInTheDocument();
        submit();
        await waitFor(() => expect(errorSpy).toHaveBeenCalledWith('Failed to save party:', expect.any(Error)));
        expect(updateMock).toHaveBeenCalledWith(expect.objectContaining({ id: 'org-1', '@type': 'Organization' }));
        errorSpy.mockRestore();
    });

    it('handles an Individual edit party with undefined sub-resource arrays', () => {
        const ind = {
            id: 'pi', '@type': 'Individual', givenName: 'A', familyName: 'B', status: 'Active',
        };
        setup(vi.fn(), vi.fn(), ind);
        renderEdit('pi');
        // All sections fall back to empty arrays -> empty-state text shown.
        expect(screen.getByText('No contact mediums added')).toBeInTheDocument();
        expect(screen.getByText('No attachments added')).toBeInTheDocument();
    });

    it('updates an Individual with all sub-resources populated', async () => {
        const updateMock = vi.fn().mockResolvedValue({});
        const ind = {
            id: 'pf', '@type': 'Individual', givenName: 'A', familyName: 'B', status: 'Active',
            contactMediums: [{ id: 'c1', mediumType: 'email', preferred: true, value: 'a@b.c' }],
            identifications: [{ id: 'i1', identificationType: 'SSN', identificationId: '1' }],
            taxExemptions: [{ id: 't1', certificateNumber: 'C', issuingJurisdiction: 'J' }],
            externalReferences: [{ id: 'e1', externalSystemId: 'S', externalReference: 'R' }],
            attachments: [{ id: 'a1', name: 'N', url: 'u', type: 'Document' }],
            relatedParties: [{ id: 'r1', role: 'x', relatedPartyId: 'rp' }],
        };
        setup(vi.fn(), updateMock, ind);
        renderEdit('pf');

        // Edit the tax-exemption date (validFor was undefined -> covers the {} fallback).
        const dateInputs = document.querySelectorAll('input[type="date"]');
        fireEvent.change(dateInputs[dateInputs.length - 1], { target: { value: '2026-06-14' } });
        submit();

        await waitFor(() => expect(updateMock).toHaveBeenCalled());
        expect(updateMock).toHaveBeenCalledWith(expect.objectContaining({
            id: 'pf',
            contactMediums: expect.any(Array),
            identifications: expect.any(Array),
            taxExemptions: expect.any(Array),
            externalReferences: expect.any(Array),
            attachments: expect.any(Array),
            relatedParties: expect.any(Array),
        }));
    });

    it('creates an Organization with all sub-resources populated', async () => {
        const createMock = vi.fn().mockResolvedValue({});
        setup(createMock);
        renderNew();
        fireEvent.click(screen.getByRole('radio', { name: /Organization/i }));
        fireEvent.change(screen.getByLabelText('Trading Name *'), { target: { value: 'Acme' } });

        fireEvent.click(screen.getByRole('button', { name: /add contact/i }));
        fireEvent.change(screen.getByPlaceholderText('email@example.com'), { target: { value: 'a@b.c' } });
        fireEvent.click(screen.getByRole('button', { name: /add id/i }));
        fireEvent.click(screen.getByRole('button', { name: /add exemption/i }));
        fireEvent.click(screen.getByRole('button', { name: /add reference/i }));
        fireEvent.click(screen.getByRole('button', { name: /add attachment/i }));
        fireEvent.click(screen.getByRole('button', { name: /add relationship/i }));
        submit();

        await waitFor(() => expect(createMock).toHaveBeenCalled());
        const payload = createMock.mock.calls[0][0];
        expect(payload.contactMediums).toHaveLength(1);
        expect(payload.identifications).toHaveLength(1);
        expect(payload.taxExemptions).toHaveLength(1);
        expect(payload.externalReferences).toHaveLength(1);
        expect(payload.attachments).toHaveLength(1);
        expect(payload.relatedParties).toHaveLength(1);
    });

    it('disables the submit button while pending', () => {
        setup(vi.fn(), vi.fn(), undefined, true);
        renderNew();
        expect(screen.getByRole('button', { name: /create party/i })).toBeDisabled();
    });

    it('shows a loading state while the party loads in edit mode', () => {
        vi.mocked(api.useParty).mockReturnValue({ data: undefined, isLoading: true } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
        vi.mocked(api.useCreateParty).mockReturnValue({ mutateAsync: vi.fn(), isPending: false } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
        vi.mocked(api.useUpdateParty).mockReturnValue({ mutateAsync: vi.fn(), isPending: false } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
        renderEdit('p9');
        expect(screen.getByText('Loading party...')).toBeInTheDocument();
    });
});
