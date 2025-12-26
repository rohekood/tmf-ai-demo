import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import PartyListPage from './PartyListPage';
import { MemoryRouter } from 'react-router-dom';
import * as api from './api';
import { type Individual, type Organization } from './types';

// Mock hooks
vi.mock('./api');

const mockParties: (Individual | Organization)[] = [
    {
        id: 'p1',
        '@type': 'Individual',
        givenName: 'John',
        familyName: 'Doe',
        status: 'active',
        identifications: []
    } as Individual,
    {
        id: 'p2',
        '@type': 'Organization',
        tradingName: 'Acme Corp',
        status: 'active',
        identifications: [{ identificationType: 'taxNr', identificationId: '123' }]
    } as Organization
];

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

        render(
            <MemoryRouter>
                <PartyListPage />
            </MemoryRouter>
        );
        expect(screen.getByRole('status')).toHaveTextContent('Loading parties...');
    });

    it('renders party list', () => {
        vi.mocked(api.useParties).mockReturnValue({
            data: mockParties,
            isLoading: false,
            error: null
        } as unknown as ReturnType<typeof api.useParties>);

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
        vi.mocked(api.useParties).mockReturnValue({
            data: [],
            isLoading: false,
            error: null
        } as unknown as ReturnType<typeof api.useParties>);

        render(
            <MemoryRouter>
                <PartyListPage />
            </MemoryRouter>
        );

        expect(screen.getByRole('status')).toHaveTextContent('No parties found.');
        expect(screen.getByRole('link', { name: /create your first party/i })).toBeInTheDocument();
    });
});
