import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import CustomerDetailPage from './CustomerDetailPage';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import * as api from './api';

vi.mock('./api');

const mockCustomer = {
    id: 'c1',
    name: 'Test Customer',
    status: 'active',
    partyId: 'p1',
    partyName: 'Test Party',
    creditProfile: {
        creditScore: 750,
        creditRiskScore: 10
    },
    accounts: [
        { id: 'a1', name: 'Main Account', accountType: 'postpaid', accountStatus: 'active' }
    ],
    privacyConsents: [],
    taxExemptions: []
};

describe('CustomerDetailPage', () => {
    beforeEach(() => {
        vi.resetAllMocks();
        (api.useDeleteCustomer as any).mockReturnValue({
            mutate: vi.fn(),
            isPending: false
        });
    });

    it('renders loading state', () => {
        (api.useCustomer as any).mockReturnValue({
            data: undefined,
            isLoading: true,
            error: null
        });

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
        (api.useCustomer as any).mockReturnValue({
            data: mockCustomer,
            isLoading: false,
            error: null
        });

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
        expect(screen.getByText('750')).toBeInTheDocument();
        expect(screen.getByText('Main Account')).toBeInTheDocument();
    });

    it('renders error state', () => {
        (api.useCustomer as any).mockReturnValue({
            data: undefined,
            isLoading: false,
            error: { message: 'Not found' }
        });

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
