import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import CustomerEditPage from './CustomerEditPage';
import { MemoryRouter, Route, Routes } from 'react-router-dom';

// Mocks
import userEvent from '@testing-library/user-event';

// Mocks
const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom');
    return {
        ...actual,
        useNavigate: () => mockNavigate,
    };
});

import * as partyApi from '../parties/api';
vi.mock('../parties/api');

const mockCustomerData = {
    id: '123',
    name: 'Test Customer',
    status: 'active',
    taxExemptions: [],
    privacyConsents: [],
    creditProfiles: [],
    accounts: []
};

const mockMutateAsync = vi.fn();
vi.mock('./api', () => ({
    useCustomer: (id: string) => ({
        data: id === '123' ? mockCustomerData : null,
        isLoading: false,
    }),
    useUpdateCustomer: () => ({
        mutateAsync: mockMutateAsync,
        isPending: false,
    }),
}));

describe('CustomerEditPage', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        // Default mock for parties
        vi.mocked(partyApi.useParties).mockReturnValue({
            data: [],
            isLoading: false
        } as unknown as ReturnType<typeof partyApi.useParties>);
    });

    it('updates customer party without sending redundant fields', async () => {
        const user = userEvent.setup();

        // Mock parties response
        vi.mocked(partyApi.useParties).mockReturnValue({
            data: [
                { id: 'p2', '@type': 'Individual', givenName: 'New', familyName: 'Party', status: 'active', identifications: [] }
            ],
            isLoading: false
        } as unknown as ReturnType<typeof partyApi.useParties>);

        render(
            <MemoryRouter initialEntries={['/customers/123/edit']}>
                <Routes>
                    <Route path="/customers/:id/edit" element={<CustomerEditPage />} />
                </Routes>
            </MemoryRouter>
        );

        // Search for new party
        // First click 'Change Party'
        const changeButton = screen.getByRole('button', { name: /change party/i });
        await user.click(changeButton);

        const searchInput = await screen.findByPlaceholderText('Search by name...');
        await user.type(searchInput, 'New');

        // Select new party
        const partyOption = await screen.findByText('New Party');
        await user.click(partyOption);

        // Submit
        const saveButton = screen.getByRole('button', { name: /save changes/i });
        await user.click(saveButton);

        await waitFor(() => {
            expect(mockMutateAsync).toHaveBeenCalledWith(expect.objectContaining({
                id: '123',
                partyId: 'p2', // Updated party ID
            }));


        });
    });

    it('renders customer data correctly', () => {
        render(
            <MemoryRouter initialEntries={['/customers/123/edit']}>
                <Routes>
                    <Route path="/customers/:id/edit" element={<CustomerEditPage />} />
                </Routes>
            </MemoryRouter>
        );

        expect(screen.getByDisplayValue('Test Customer')).toBeInTheDocument();
        expect(screen.getByDisplayValue('Active')).toBeInTheDocument();
    });

    it('submits form with updated data including new fields', async () => {
        const user = userEvent.setup();
        render(
            <MemoryRouter initialEntries={['/customers/123/edit']}>
                <Routes>
                    <Route path="/customers/:id/edit" element={<CustomerEditPage />} />
                </Routes>
            </MemoryRouter>
        );

        // Update Name
        const nameInput = screen.getByLabelText('Customer Name');
        await user.clear(nameInput);
        await user.type(nameInput, 'Updated Name');

        // Add Credit Profile
        await user.click(screen.getByRole('button', { name: 'Add Profile' }));
        const riskInput = screen.getByLabelText('Credit Risk Score');
        await user.type(riskInput, '750');

        // Add Account
        await user.click(screen.getByRole('button', { name: 'Add Account' }));
        const accNameInput = screen.getByLabelText('Account Name');
        await user.type(accNameInput, 'New Account');

        // Submit
        const saveButton = screen.getByRole('button', { name: /save changes/i });
        await user.click(saveButton);

        await waitFor(() => {
            expect(mockMutateAsync).toHaveBeenCalledWith(expect.objectContaining({
                id: '123',
                name: 'Updated Name',
                creditProfiles: expect.arrayContaining([
                    expect.objectContaining({ creditRiskScore: 750 })
                ]),
                accounts: expect.arrayContaining([
                    expect.objectContaining({ name: 'New Account' })
                ])
            }));
        });

        expect(mockNavigate).toHaveBeenCalledWith('/customers/123');
    });

    it('adds a tax exemption', () => {
        render(
            <MemoryRouter initialEntries={['/customers/123/edit']}>
                <Routes>
                    <Route path="/customers/:id/edit" element={<CustomerEditPage />} />
                </Routes>
            </MemoryRouter>
        );

        const addButton = screen.getByRole('button', { name: /add exemption/i });
        fireEvent.click(addButton);

        expect(screen.getByPlaceholderText('Certificate #')).toBeInTheDocument();
        expect(screen.getByPlaceholderText('e.g., US-CA')).toBeInTheDocument();
    });
});
