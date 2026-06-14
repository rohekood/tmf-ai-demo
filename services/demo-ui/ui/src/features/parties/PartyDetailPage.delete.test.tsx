import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import PartyDetailPage from './PartyDetailPage';
import * as api from './api';
import { NotificationProvider } from '../../design-system/components/common/ToastProvider';
import type { Individual, Organization } from './types';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
    return { ...actual, useNavigate: () => mockNavigate };
});

const getMock = vi.fn();
vi.mock('../../api/client', () => ({ apiClient: { get: (...args: unknown[]) => getMock(...args) } }));

vi.mock('./api', async () => {
    const actual = await vi.importActual('./api');
    return { ...actual, useParty: vi.fn(), useDeleteParty: vi.fn() };
});

const individual: Individual = {
    id: 'p1', '@type': 'Individual', givenName: 'John', familyName: 'Doe',
    middleName: 'M', birthDate: '1990-01-15', gender: 'male', status: 'Active',
    contactMediums: [
        { id: 'c1', mediumType: 'email', value: 'john@example.com', preferred: true },
        { id: 'c2', mediumType: 'phone', value: '+372 1', preferred: false },
        { id: 'c3', mediumType: 'postal', preferred: false, street1: 'Main 1', city: 'Tallinn', country: 'EE' },
    ],
    identifications: [{ id: 'i1', identificationType: 'Passport', identificationId: 'X1', issuingAuthority: 'PPA' }],
    relatedParties: [{ id: 'r1', role: 'guardian', relatedPartyId: 'rp1', relatedPartyName: 'Jane', permissions: ['read', 'write'] }],
};

const organization: Organization = {
    id: 'p2', '@type': 'Organization', tradingName: 'Acme', isLegalEntity: false,
    organizationType: 'LLC', status: 'Active', contactMediums: [], identifications: [], relatedParties: [],
};

function mutateImpl(behavior: 'success' | 'error') {
    return (_id: string, opts: { onSuccess: () => void; onError: (e: Error) => void }) => {
        if (behavior === 'success') opts.onSuccess();
        else opts.onError(new Error('explode'));
    };
}

function setup(party: unknown, mutate: (...args: never[]) => void = vi.fn()) {
    vi.mocked(api.useParty).mockReturnValue({ data: party, isLoading: false, error: null } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
    vi.mocked(api.useDeleteParty).mockReturnValue({ mutate, isPending: false } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
}

const renderPage = (route = '/parties/p1') =>
    render(
        <MemoryRouter initialEntries={[route]}>
            <NotificationProvider>
                <Routes>
                    <Route path="/parties/:id" element={<PartyDetailPage />} />
                </Routes>
            </NotificationProvider>
        </MemoryRouter>
    );

beforeEach(() => {
    mockNavigate.mockClear();
    getMock.mockReset();
    window.confirm = vi.fn(() => true);
});

describe('PartyDetailPage rendering branches', () => {
    it('renders all Individual detail branches', () => {
        setup(individual);
        renderPage();
        expect(screen.getByText('15.01.1990')).toBeInTheDocument(); // Estonian birth date
        expect(screen.getByText('male')).toBeInTheDocument();
        expect(screen.getByText('+372 1')).toBeInTheDocument();
        expect(screen.getByText('Main 1, Tallinn, EE')).toBeInTheDocument();
        expect(screen.getByText('Issued by: PPA')).toBeInTheDocument();
        expect(screen.getByText('Jane')).toBeInTheDocument();
        expect(screen.getByText(/read, write/)).toBeInTheDocument();
    });

    it('renders Organization detail branches', () => {
        setup(organization);
        renderPage('/parties/p2');
        expect(screen.getAllByText('Acme').length).toBeGreaterThan(0);
        expect(screen.getByText('No')).toBeInTheDocument(); // legal entity false
        expect(screen.getByText('LLC')).toBeInTheDocument();
    });

    it('renders an error state', () => {
        vi.mocked(api.useParty).mockReturnValue({ data: undefined, isLoading: false, error: new Error('nope') } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
        vi.mocked(api.useDeleteParty).mockReturnValue({ mutate: vi.fn(), isPending: false } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
        renderPage();
        expect(screen.getByText(/Failed to load party: nope/)).toBeInTheDocument();
    });
});

describe('PartyDetailPage deletion polling', () => {
    beforeEach(() => vi.useFakeTimers());
    afterEach(() => vi.useRealTimers());

    it('navigates away once the party reports Deleted', async () => {
        getMock.mockResolvedValue({ data: { status: 'Deleted' } });
        setup(individual, mutateImpl('success'));
        renderPage();

        fireEvent.click(screen.getByRole('button', { name: /delete/i }));
        await vi.advanceTimersByTimeAsync(600);
        expect(mockNavigate).toHaveBeenCalledWith('/parties');
    });

    it('stays on the page when the party is still Active', async () => {
        getMock.mockResolvedValue({ data: { status: 'Active' } });
        setup(individual, mutateImpl('success'));
        renderPage();

        fireEvent.click(screen.getByRole('button', { name: /delete/i }));
        await vi.advanceTimersByTimeAsync(600);
        expect(mockNavigate).not.toHaveBeenCalled();
    });

    it('keeps polling while DeletionPending then resolves', async () => {
        getMock
            .mockResolvedValueOnce({ data: { status: 'DeletionPending' } })
            .mockResolvedValueOnce({ data: { status: 'Deleted' } });
        setup(individual, mutateImpl('success'));
        renderPage();

        fireEvent.click(screen.getByRole('button', { name: /delete/i }));
        await vi.advanceTimersByTimeAsync(600); // first poll -> DeletionPending
        await vi.advanceTimersByTimeAsync(1000); // second poll -> Deleted
        expect(mockNavigate).toHaveBeenCalledWith('/parties');
    });

    it('navigates for an unexpected status', async () => {
        getMock.mockResolvedValue({ data: { status: 'Archived' } });
        setup(individual, mutateImpl('success'));
        renderPage();
        fireEvent.click(screen.getByRole('button', { name: /delete/i }));
        await vi.advanceTimersByTimeAsync(600);
        expect(mockNavigate).toHaveBeenCalledWith('/parties');
    });

    it('treats a 404 as a successful deletion', async () => {
        getMock.mockRejectedValue({ response: { status: 404 } });
        setup(individual, mutateImpl('success'));
        renderPage();
        fireEvent.click(screen.getByRole('button', { name: /delete/i }));
        await vi.advanceTimersByTimeAsync(600);
        expect(mockNavigate).toHaveBeenCalledWith('/parties');
    });

    it('logs and surfaces a non-404 status-check error', async () => {
        const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
        getMock.mockRejectedValue(new Error('network'));
        setup(individual, mutateImpl('success'));
        renderPage();
        fireEvent.click(screen.getByRole('button', { name: /delete/i }));
        await vi.advanceTimersByTimeAsync(600);
        expect(errorSpy).toHaveBeenCalledWith('Error checking status', expect.any(Error));
        expect(mockNavigate).not.toHaveBeenCalled();
        errorSpy.mockRestore();
    });

    it('times out after repeated DeletionPending polls', async () => {
        getMock.mockResolvedValue({ data: { status: 'DeletionPending' } });
        setup(individual, mutateImpl('success'));
        renderPage();
        fireEvent.click(screen.getByRole('button', { name: /delete/i }));
        await vi.advanceTimersByTimeAsync(600);
        for (let i = 0; i < 17; i++) {
            await vi.advanceTimersByTimeAsync(1000);
        }
        expect(mockNavigate).toHaveBeenCalledWith('/parties');
    });

    it('surfaces a delete mutation error', async () => {
        setup(individual, mutateImpl('error'));
        renderPage();
        fireEvent.click(screen.getByRole('button', { name: /delete/i }));
        // onError path runs synchronously; no polling expected
        expect(getMock).not.toHaveBeenCalled();
        expect(mockNavigate).not.toHaveBeenCalled();
    });
});
