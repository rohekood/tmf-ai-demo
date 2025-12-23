import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import CustomerListPage from './CustomerListPage';
import { MemoryRouter } from 'react-router-dom';
import * as api from './api';

// Mock the hooks
vi.mock('./api');

const mockCustomers = [
    {
        id: '1',
        name: 'John Doe',
        status: 'active',
        partyId: 'p1',
        partyName: 'John Party',
        accounts: []
    },
    {
        id: '2',
        name: 'Jane Smith',
        status: 'prospecting',
        partyId: 'p2',
        partyName: 'Jane Party',
        accounts: [{ id: 'a1' }]
    }
];

describe('CustomerListPage', () => {
    beforeEach(() => {
        vi.resetAllMocks();
        (api.useDeleteCustomer as any).mockReturnValue({
            mutate: vi.fn(),
            isPending: false
        });
    });

    it('renders loading state', () => {
        (api.useCustomers as any).mockReturnValue({
            data: undefined,
            isLoading: true,
            error: null
        });

        render(
            <MemoryRouter>
                <CustomerListPage />
            </MemoryRouter>
        );
        expect(screen.getByRole('status')).toHaveTextContent('Loading customers...');
    });

    it('renders customer list', () => {
        (api.useCustomers as any).mockReturnValue({
            data: mockCustomers,
            isLoading: false,
            error: null
        });

        render(
            <MemoryRouter>
                <CustomerListPage />
            </MemoryRouter>
        );

        // Verify headers
        expect(screen.getByRole('columnheader', { name: 'Name' })).toBeInTheDocument();
        expect(screen.getByRole('columnheader', { name: 'Status' })).toBeInTheDocument();

        // Verify data cells
        expect(screen.getByRole('cell', { name: 'John Doe' })).toBeInTheDocument();
        expect(screen.getByRole('cell', { name: 'Jane Smith' })).toBeInTheDocument();

        // Count rows (including header row)
        expect(screen.getAllByRole('row')).toHaveLength(3);
    });

    it('renders empty state', () => {
        (api.useCustomers as any).mockReturnValue({
            data: [],
            isLoading: false,
            error: null
        });

        render(
            <MemoryRouter>
                <CustomerListPage />
            </MemoryRouter>
        );

        expect(screen.getByRole('status')).toHaveTextContent('No customers found.');
        expect(screen.getByRole('link', { name: /onboard your first customer/i })).toBeInTheDocument();
    });
});
