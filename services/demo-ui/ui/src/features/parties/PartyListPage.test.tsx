import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import PartyListPage from './PartyListPage';
import { MemoryRouter } from 'react-router-dom';
import * as api from './api';

// Mock hooks
vi.mock('./api');

const mockParties = [
    {
        id: 'p1',
        '@type': 'Individual',
        givenName: 'John',
        familyName: 'Doe',
        status: 'Active',
        identifications: []
    },
    {
        id: 'p2',
        '@type': 'Organization',
        tradingName: 'Acme Corp',
        status: 'Active',
        identifications: [{ identificationType: 'taxNr', identificationId: '123' }]
    }
];

describe('PartyListPage', () => {
    beforeEach(() => {
        vi.resetAllMocks();
        (api.useDeleteParty as any).mockReturnValue({
            mutate: vi.fn(),
            isPending: false
        });
    });

    it('renders loading state', () => {
        (api.useParties as any).mockReturnValue({
            data: undefined,
            isLoading: true,
            error: null
        });

        render(
            <MemoryRouter>
                <PartyListPage />
            </MemoryRouter>
        );
        expect(screen.getByRole('status')).toHaveTextContent('Loading parties...');
    });

    it('renders party list', () => {
        (api.useParties as any).mockReturnValue({
            data: mockParties,
            isLoading: false,
            error: null
        });

        render(
            <MemoryRouter>
                <PartyListPage />
            </MemoryRouter>
        );


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
        (api.useParties as any).mockReturnValue({
            data: [],
            isLoading: false,
            error: null
        });

        render(
            <MemoryRouter>
                <PartyListPage />
            </MemoryRouter>
        );

        expect(screen.getByRole('status')).toHaveTextContent('No parties found.');
        expect(screen.getByRole('link', { name: /create your first party/i })).toBeInTheDocument();
    });
});
