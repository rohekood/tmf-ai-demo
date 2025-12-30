import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import CustomerDetailPage from './CustomerDetailPage';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import * as api from './api';
import { type Customer } from './types';

vi.mock('./api');

const mockCustomer: Customer = {
    id: 'c1',
    name: 'Test Customer',
    status: 'active',
    partyId: 'p1',
    partyName: 'Test Party',

    privacyConsents: [],
    creditProfiles: [{
        id: 'cp1',
        creditScore: 750,
        creditRiskScore: 10
    }],
    accounts: [
        { id: 'a1', name: 'Main Account', accountType: 'postpaid', accountStatus: 'active' }
    ],
    characteristics: [],
    contactMediums: []
};

describe('CustomerDetailPage', () => {
    beforeEach(() => {
        vi.resetAllMocks();
        vi.mocked(api.useDeleteCustomer).mockReturnValue({
            mutate: vi.fn(),
            isPending: false
        } as unknown as ReturnType<typeof api.useDeleteCustomer>);
    });

    it('renders loading state', () => {
        vi.mocked(api.useCustomer).mockReturnValue({
            data: undefined,
            isLoading: true,
            error: null
        } as unknown as ReturnType<typeof api.useCustomer>);

        render(
            <MemoryRouter initialEntries={['/customers/c1']}>
                <Routes>
                    <Route path="/customers/:id" element={<CustomerDetailPage />} />
                </Routes>
            </MemoryRouter>
        );

        expect(screen.getByRole('status')).toHaveTextContent('Loading customer details...');
    });

    it('renders customer details', () => {
        vi.mocked(api.useCustomer).mockReturnValue({
            data: mockCustomer,
            isLoading: false,
            error: null
        } as unknown as ReturnType<typeof api.useCustomer>);

        render(
            <MemoryRouter initialEntries={['/customers/c1']}>
                <Routes>
                    <Route path="/customers/:id" element={<CustomerDetailPage />} />
                </Routes>
            </MemoryRouter>
        );

        expect(screen.getByRole('heading', { name: 'Test Customer' })).toBeInTheDocument();
        expect(screen.getByRole('heading', { name: 'Basic Information' })).toBeInTheDocument();
        // Check for specific term/definitions or just text presence if roles are too generic
        expect(screen.getByText('c1')).toBeInTheDocument();

        // Assert Credit Profile present
        expect(screen.getByText('750')).toBeInTheDocument();

        // Assert Accounts present
        expect(screen.getByText('Main Account')).toBeInTheDocument();

        // Regression Test: Check Edit Link URL
        const editLink = screen.getByRole('link', { name: /edit/i });
        expect(editLink).toHaveAttribute('href', '/customers/c1/edit');

        // Regression Test: Check Party Link URL (Top META)
        // We look for the link that contains "Party:" text or verify strictly by href if identifiable
        const partyLink = screen.getByRole('link', { name: /Party:/i });
        expect(partyLink).toHaveAttribute('href', '/parties/p1');
    });

    it('renders error state', () => {
        vi.mocked(api.useCustomer).mockReturnValue({
            data: undefined,
            isLoading: false,
            error: { message: 'Not found', name: 'Error' }
        } as unknown as ReturnType<typeof api.useCustomer>);

        render(
            <MemoryRouter initialEntries={['/customers/c1']}>
                <Routes>
                    <Route path="/customers/:id" element={<CustomerDetailPage />} />
                </Routes>
            </MemoryRouter>
        );

        expect(screen.getByRole('alert')).toHaveTextContent('Failed to load customer: Not found');
    });
});
