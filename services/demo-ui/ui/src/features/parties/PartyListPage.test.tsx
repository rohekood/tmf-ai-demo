import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import PartyListPage from './PartyListPage';
import { MemoryRouter } from 'react-router-dom';
import * as api from './api';
import { type Individual, type Organization } from './types';
import { NotificationProvider } from '../../components/common/ToastProvider';

// Mock hooks
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

const mockParties: (Individual | Organization)[] = [
    {
        id: 'p1',
        '@type': 'Individual',
        givenName: 'John',
        familyName: 'Doe',
        status: 'Active',
        identifications: []
    } as Individual,
    {
        id: 'p2',
        '@type': 'Organization',
        tradingName: 'Acme Corp',
        status: 'Active',
        isLegalEntity: true,
        identifications: [{ identificationType: 'taxNr', identificationId: '123' }]
    } as Organization
];

// Helper to render with all necessary providers
const renderWithProviders = (ui: React.ReactElement) => {
    return render(
        <MemoryRouter>
            <NotificationProvider>
                {ui}
            </NotificationProvider>
        </MemoryRouter>
    );
};

describe('PartyListPage', () => {
    beforeEach(() => {
        vi.resetAllMocks();
        vi.mocked(api.useDeleteParty).mockReturnValue({
            mutate: vi.fn(),
            isPending: false
        } as unknown as ReturnType<typeof api.useDeleteParty>);
    });



    it('renders loading state', () => {
        vi.mocked(api.useParties).mockReturnValue({
            data: undefined,
            isLoading: true,
            error: null
        } as unknown as ReturnType<typeof api.useParties>);

        renderWithProviders(<PartyListPage />);
        expect(screen.getByRole('status')).toHaveTextContent('Loading parties...');
    });

    it('renders party list', () => {
        vi.mocked(api.useParties).mockReturnValue({
            data: mockParties,
            isLoading: false,
            error: null
        } as unknown as ReturnType<typeof api.useParties>);

        renderWithProviders(<PartyListPage />);

        // Headers
        expect(screen.getByRole('columnheader', { name: 'Name' })).toBeInTheDocument();
        expect(screen.getByRole('columnheader', { name: 'Type' })).toBeInTheDocument();

        // Check for Type badges (accessible as buttons)
        expect(screen.getByRole('button', { name: /individual/i })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /organization/i })).toBeInTheDocument();

        // Cells
        expect(screen.getByRole('cell', { name: 'John Doe' })).toBeInTheDocument();
        expect(screen.getByRole('cell', { name: 'Acme Corp' })).toBeInTheDocument();

        // Rows (header + 2 items)
        expect(screen.getAllByRole('row')).toHaveLength(3);
    });

    it('renders empty state', () => {
        vi.mocked(api.useParties).mockReturnValue({
            data: [],
            isLoading: false,
            error: null
        } as unknown as ReturnType<typeof api.useParties>);

        renderWithProviders(<PartyListPage />);

        expect(screen.getByRole('status')).toHaveTextContent('No parties found.');
        expect(screen.getByRole('link', { name: /create your first party/i })).toBeInTheDocument();
    });

    it('initiates deletion when delete button clicked', async () => {
        const user = userEvent.setup();
        const mutateMock = vi.fn();

        vi.mocked(api.useDeleteParty).mockReturnValue({
            mutate: mutateMock,
            isPending: false
        } as unknown as ReturnType<typeof api.useDeleteParty>);

        vi.mocked(api.useParties).mockReturnValue({
            data: mockParties,
            isLoading: false,
            error: null
        } as unknown as ReturnType<typeof api.useParties>);

        window.confirm = vi.fn(() => true);

        renderWithProviders(<PartyListPage />);

        const deleteBtn = screen.getAllByTitle('Delete')[0];
        await user.click(deleteBtn);

        expect(window.confirm).toHaveBeenCalled();
        expect(mutateMock).toHaveBeenCalledWith('p1', expect.any(Object));
    });

    it('does not delete when confirm is cancelled', async () => {
        const user = userEvent.setup();
        const mutateMock = vi.fn();

        vi.mocked(api.useDeleteParty).mockReturnValue({
            mutate: mutateMock,
            isPending: false
        } as unknown as ReturnType<typeof api.useDeleteParty>);

        vi.mocked(api.useParties).mockReturnValue({
            data: mockParties,
            isLoading: false,
            error: null
        } as unknown as ReturnType<typeof api.useParties>);

        window.confirm = vi.fn(() => false);

        renderWithProviders(<PartyListPage />);

        const deleteBtn = screen.getAllByTitle('Delete')[0];
        await user.click(deleteBtn);

        expect(window.confirm).toHaveBeenCalled();
        expect(mutateMock).not.toHaveBeenCalled();
    });
});
