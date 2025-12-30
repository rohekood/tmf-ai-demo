import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import CustomerOnboardPage from './CustomerOnboardPage';
import { MemoryRouter } from 'react-router-dom';
import * as api from './api';
import * as partyApi from '../parties/api';
import { type Individual, type Organization } from '../parties/types';

// Mock hooks
vi.mock('./api');
vi.mock('../parties/api');
const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom');
    return {
        ...actual,
        useNavigate: () => mockNavigate,
    };
});

const mockParties: (Individual | Organization)[] = [
    {
        id: 'p1',
        givenName: 'John',
        familyName: 'Doe',
        '@type': 'Individual',
        status: 'Active',
        identifications: []
    } as Individual,
    {
        id: 'p2',
        tradingName: 'Acme Corp',
        '@type': 'Organization',
        status: 'Active',
        isLegalEntity: true,
        identifications: []
    } as Organization
];

describe('CustomerOnboardPage', () => {
    let mutateAsyncMock: ReturnType<typeof vi.fn>;

    beforeEach(() => {
        vi.resetAllMocks();
        mutateAsyncMock = vi.fn();
        vi.mocked(api.useOnboardCustomer).mockReturnValue({
            mutateAsync: mutateAsyncMock,
            isPending: false
        } as unknown as ReturnType<typeof api.useOnboardCustomer>);
        vi.mocked(partyApi.useParties).mockReturnValue({
            data: [],
            isLoading: false
        } as unknown as ReturnType<typeof partyApi.useParties>);
    });

    it('renders correctly', () => {
        render(
            <MemoryRouter>
                <CustomerOnboardPage />
            </MemoryRouter>
        );
        expect(screen.getByRole('heading', { name: 'Onboard Customer' })).toBeInTheDocument();
        expect(screen.getByText('Select Party')).toBeInTheDocument();
    });

    it('handles party search and selection', async () => {
        const user = userEvent.setup();
        vi.mocked(partyApi.useParties).mockReturnValue({
            data: mockParties,
            isLoading: false
        } as unknown as ReturnType<typeof partyApi.useParties>);

        render(
            <MemoryRouter>
                <CustomerOnboardPage />
            </MemoryRouter>
        );

        const searchInput = screen.getByPlaceholderText('Search by name...');
        await user.type(searchInput, 'John');

        // Wait/Check for parties list item
        const partyOption = await screen.findByRole('button', { name: /John Doe/i });
        expect(partyOption).toBeInTheDocument();

        // Select logic
        await user.click(partyOption);

        expect(await screen.findByRole('button', { name: 'Change' })).toBeInTheDocument();
    });

    it('prefills customer name from party', async () => {
        const user = userEvent.setup();
        vi.mocked(partyApi.useParties).mockReturnValue({
            data: mockParties,
            isLoading: false
        } as unknown as ReturnType<typeof partyApi.useParties>);

        render(
            <MemoryRouter>
                <CustomerOnboardPage />
            </MemoryRouter>
        );

        // Search for party
        const searchInput = screen.getByPlaceholderText('Search by name...');
        await user.type(searchInput, 'John');

        // Select party "John Doe"
        await user.click(await screen.findByText('John Doe'));

        // Verify Name field is prefilled
        const nameInput = screen.getByLabelText('Customer Name') as HTMLInputElement;
        expect(nameInput.value).toBe('John Doe');
    });

    it('submits form with all optional fields', async () => {
        const user = userEvent.setup();
        vi.mocked(partyApi.useParties).mockReturnValue({
            data: mockParties,
            isLoading: false
        } as unknown as ReturnType<typeof partyApi.useParties>);

        render(
            <MemoryRouter>
                <CustomerOnboardPage />
            </MemoryRouter>
        );

        // Select party
        const searchInput = screen.getByPlaceholderText('Search by name...');
        await user.type(searchInput, 'John');
        await user.click(await screen.findByText('John Doe'));

        // Enter name (clear first because it's prefilled)
        const nameInput = screen.getByLabelText('Customer Name');
        await user.clear(nameInput);
        await user.type(nameInput, 'Full Customer');

        // Add Credit Profile
        await user.click(screen.getByLabelText('Add Credit Profile'));
        const riskInput = screen.getByLabelText('Risk Score');
        await user.type(riskInput, '850');
        const scoreInput = screen.getByLabelText('Credit Score');
        await user.type(scoreInput, '100');

        // Add Account
        await user.click(screen.getByRole('button', { name: 'Add Account' }));
        const accNameInput = screen.getByPlaceholderText('Account Name');
        await user.type(accNameInput, 'Primary Checking');
        const accTypeInput = screen.getByPlaceholderText('Type');
        await user.type(accTypeInput, 'Checking');

        // Add Tax Exemption
        await user.click(screen.getByRole('button', { name: 'Add Exemption' }));
        const certInput = screen.getByPlaceholderText('Certificate Number');
        await user.type(certInput, 'TAX-123');
        const jurInput = screen.getByPlaceholderText('Jurisdiction');
        await user.type(jurInput, 'CA');

        // Submit
        const submitBtn = screen.getByRole('button', { name: /onboard customer/i });
        await user.click(submitBtn);

        await waitFor(() => {
            expect(mutateAsyncMock).toHaveBeenCalledWith(expect.objectContaining({
                name: 'Full Customer',
                partyId: 'p1',
                accounts: [
                    expect.objectContaining({
                        name: 'Primary Checking',
                        accountType: 'Checking',
                        accountStatus: 'active',
                        billFormat: 'PDF',
                        billingCycle: 'Monthly'
                    })
                ],
                // Verify arrays are empty as expected
                relatedParties: [],
                paymentMethods: [],
                marketSegments: [],
                appliedBillingRates: []
            }));
            expect(mockNavigate).toHaveBeenCalledWith('/customers');
        });
    });
});
