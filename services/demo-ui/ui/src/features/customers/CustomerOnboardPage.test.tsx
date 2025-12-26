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
        status: 'active',
        identifications: []
    } as Individual,
    {
        id: 'p2',
        tradingName: 'Acme Corp',
        '@type': 'Organization',
        status: 'active',
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

        const searchInput = screen.getByPlaceholderText('Search parties by name...');
        await user.type(searchInput, 'John');

        // Wait/Check for parties list item
        const partyOption = await screen.findByRole('button', { name: /John Doe/i });
        expect(partyOption).toBeInTheDocument();

        // Select logic
        await user.click(partyOption);

        expect(await screen.findByRole('button', { name: 'Change' })).toBeInTheDocument();
    });

    it('submits form when valid', async () => {
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
        await user.click(screen.getByText('John Doe'));

        // Enter name
        const nameInput = screen.getByLabelText('Customer Name *');
        await user.type(nameInput, 'New Customer');

        // Submit
        const submitBtn = screen.getByRole('button', { name: /onboard customer/i });
        expect(submitBtn).toBeEnabled();

        await user.click(submitBtn);

        await waitFor(() => {
            expect(mutateAsyncMock).toHaveBeenCalledWith({
                name: 'New Customer',
                partyId: 'p1'
            });
            expect(mockNavigate).toHaveBeenCalledWith('/customers');
        });
    });
});
