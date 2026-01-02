import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import PartyDetailPage from './PartyDetailPage';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import * as api from './api';
import { type Individual, type Organization } from './types';
import { NotificationProvider } from '../../design-system/components/common/ToastProvider';

vi.mock('./api');
vi.mock('@tanstack/react-query', async () => {
    const actual = await vi.importActual('@tanstack/react-query');
    return {
        ...actual,
        useQueryClient: vi.fn(() => ({
            invalidateQueries: vi.fn()
        })),
    };
});

const mockIndividual: Individual = {
    id: 'p1',
    '@type': 'Individual',
    givenName: 'John',
    familyName: 'Doe',
    status: 'Active',
    contactMediums: [{ id: 'cm1', mediumType: 'email', value: 'john@example.com', preferred: true }],
    identifications: [],
    relatedParties: []
};

const mockOrganization: Organization = {
    id: 'p2',
    '@type': 'Organization',
    tradingName: 'Acme Corp',
    isLegalEntity: true,
    status: 'Active',
    contactMediums: [],
    identifications: [],
    relatedParties: []
};

// Helper to render with all necessary providers
const renderWithProviders = (ui: React.ReactElement, { route = '/parties/p1' } = {}) => {
    return render(
        <MemoryRouter initialEntries={[route]}>
            <NotificationProvider>
                <Routes>
                    <Route path="/parties/:id" element={ui} />
                    <Route path="/parties" element={<div>Parties List</div>} />
                </Routes>
            </NotificationProvider>
        </MemoryRouter>
    );
};

describe('PartyDetailPage', () => {
    beforeEach(() => {
        vi.resetAllMocks();
        vi.mocked(api.useDeleteParty).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
        } as unknown as ReturnType<typeof api.useDeleteParty>);
    });

    it('renders loading state', () => {
        vi.mocked(api.useParty).mockReturnValue({
            data: undefined,
            isLoading: true,
            error: null,
        } as unknown as ReturnType<typeof api.useParty>);

        renderWithProviders(<PartyDetailPage />);

        expect(screen.getByRole('status')).toHaveTextContent('Loading party details...');
    });

    it('renders individual details', () => {
        vi.mocked(api.useParty).mockReturnValue({
            data: mockIndividual,
            isLoading: false,
            error: null,
        } as unknown as ReturnType<typeof api.useParty>);

        renderWithProviders(<PartyDetailPage />);

        expect(screen.getByRole('heading', { name: 'John Doe' })).toBeInTheDocument();
        expect(screen.getAllByText('Individual')[0]).toBeVisible();
        expect(screen.getByText('john@example.com')).toBeInTheDocument();
    });

    it('renders organization details', () => {
        vi.mocked(api.useParty).mockReturnValue({
            data: mockOrganization,
            isLoading: false,
            error: null,
        } as unknown as ReturnType<typeof api.useParty>);

        renderWithProviders(<PartyDetailPage />, { route: '/parties/p2' });

        expect(screen.getByRole('heading', { name: 'Acme Corp' })).toBeInTheDocument();
        expect(screen.getAllByText('Organization')[0]).toBeVisible();
        expect(screen.getByText('Yes')).toBeInTheDocument();
    });

    it('initiates deletion when delete button clicked and confirmed', async () => {
        const user = userEvent.setup();
        const mutateMock = vi.fn();

        vi.mocked(api.useDeleteParty).mockReturnValue({
            mutate: mutateMock,
            isPending: false,
        } as unknown as ReturnType<typeof api.useDeleteParty>);

        vi.mocked(api.useParty).mockReturnValue({
            data: mockIndividual,
            isLoading: false,
            error: null,
        } as unknown as ReturnType<typeof api.useParty>);

        window.confirm = vi.fn(() => true);

        renderWithProviders(<PartyDetailPage />);

        const deleteBtn = screen.getByRole('button', { name: /delete/i });
        await user.click(deleteBtn);

        expect(window.confirm).toHaveBeenCalledWith('Are you sure you want to delete "John Doe"?');
        expect(mutateMock).toHaveBeenCalledWith('p1', expect.any(Object));
    });

    it('does not delete when confirm is cancelled', async () => {
        const user = userEvent.setup();
        const mutateMock = vi.fn();

        vi.mocked(api.useDeleteParty).mockReturnValue({
            mutate: mutateMock,
            isPending: false,
        } as unknown as ReturnType<typeof api.useDeleteParty>);

        vi.mocked(api.useParty).mockReturnValue({
            data: mockIndividual,
            isLoading: false,
            error: null,
        } as unknown as ReturnType<typeof api.useParty>);

        window.confirm = vi.fn(() => false);

        renderWithProviders(<PartyDetailPage />);

        const deleteBtn = screen.getByRole('button', { name: /delete/i });
        await user.click(deleteBtn);

        expect(window.confirm).toHaveBeenCalled();
        expect(mutateMock).not.toHaveBeenCalled();
    });

    it('disables delete button when deletion is pending', () => {
        vi.mocked(api.useDeleteParty).mockReturnValue({
            mutate: vi.fn(),
            isPending: true,
        } as unknown as ReturnType<typeof api.useDeleteParty>);

        vi.mocked(api.useParty).mockReturnValue({
            data: mockIndividual,
            isLoading: false,
            error: null,
        } as unknown as ReturnType<typeof api.useParty>);

        renderWithProviders(<PartyDetailPage />);

        const deleteBtn = screen.getByRole('button', { name: /delete/i });
        expect(deleteBtn).toBeDisabled();
    });
});
