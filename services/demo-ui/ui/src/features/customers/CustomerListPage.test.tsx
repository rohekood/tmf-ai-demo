import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import CustomerListPage from './CustomerListPage';
import { MemoryRouter } from 'react-router-dom';
import * as api from './api';
import { type Customer } from './types';

// Mock the hooks
vi.mock('./api');

const mockCustomers: Customer[] = [
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
        accounts: [{
            id: 'a1',
            name: 'Primary Account',
            accountStatus: 'active',
            accountType: 'postpaid'
        }]
    }
];

describe('CustomerListPage', () => {
    beforeEach(() => {
        vi.resetAllMocks();
        vi.mocked(api.useDeleteCustomer).mockReturnValue({
            mutate: vi.fn(),
            isPending: false
        } as unknown as ReturnType<typeof api.useDeleteCustomer>);
    });

    it('renders loading state', () => {
        vi.mocked(api.useCustomers).mockReturnValue({
            data: undefined,
            isLoading: true,
            error: null
        } as unknown as ReturnType<typeof api.useCustomers>);

        render(
            <MemoryRouter>
                <CustomerListPage />
            </MemoryRouter>
        );
        expect(screen.getByRole('status')).toHaveTextContent('Loading customers...');
    });

    it('renders customer list', () => {
        vi.mocked(api.useCustomers).mockReturnValue({
            data: mockCustomers,
            isLoading: false,
            error: null
        } as unknown as ReturnType<typeof api.useCustomers>);

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

    it('handles mixed-case status values correctly', () => {
        const mixedCaseCustomers: Customer[] = [
            {
                id: '3',
                name: 'Mixed Case Customer',
                status: 'Active' as any, // Simulate backend returning capitalized value
                partyId: 'p3',
                partyName: 'Mixed Party',
                accounts: []
            }
        ];

        vi.mocked(api.useCustomers).mockReturnValue({
            data: mixedCaseCustomers,
            isLoading: false,
            error: null
        } as unknown as ReturnType<typeof api.useCustomers>);

        render(
            <MemoryRouter>
                <CustomerListPage />
            </MemoryRouter>
        );

        // Verify that the class name is lowercased
        const statusBadge = screen.getByText('Active');
        expect(statusBadge).toHaveClass('status-badge', 'active');
        expect(statusBadge).not.toHaveClass('Active');
    });

    it('renders empty state', () => {
        vi.mocked(api.useCustomers).mockReturnValue({
            data: [],
            isLoading: false,
            error: null
        } as unknown as ReturnType<typeof api.useCustomers>);

        render(
            <MemoryRouter>
                <CustomerListPage />
            </MemoryRouter>
        );

        expect(screen.getByRole('status')).toHaveTextContent('No customers found.');
        expect(screen.getByRole('link', { name: /onboard your first customer/i })).toBeInTheDocument();
    });
});
