import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import CustomerOnboardPage from './CustomerOnboardPage';
import { MemoryRouter } from 'react-router-dom';
import * as api from './api';
import * as partyApi from '../parties/api';

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

const mockParties = [
    {
        id: 'p1',
        givenName: 'John',
        familyName: 'Doe',
        '@type': 'Individual'
    },
    {
        id: 'p2',
        tradingName: 'Acme Corp',
        '@type': 'Organization'
    }
];

describe('CustomerOnboardPage', () => {
    let mutateAsyncMock: any;

    beforeEach(() => {
        vi.resetAllMocks();
        mutateAsyncMock = vi.fn();
        (api.useOnboardCustomer as any).mockReturnValue({
            mutateAsync: mutateAsyncMock,
            isPending: false
        });
        (partyApi.useParties as any).mockReturnValue({
            data: [],
            isLoading: false
        });
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
        (partyApi.useParties as any).mockReturnValue({
            data: mockParties,
            isLoading: false
        });

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
        (partyApi.useParties as any).mockReturnValue({
            data: mockParties,
            isLoading: false
        });

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
