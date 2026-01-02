import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import PartySelector from './PartySelector';
import * as api from './api';
import { type Individual, type Organization } from './types';

// Mock the API hook
vi.mock('./api', () => ({
    useParties: vi.fn(),
    // Add other exports if needed or rely on auto-mocking if mixed
}));

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

describe('PartySelector', () => {
    const mockOnSelect = vi.fn();

    beforeEach(() => {
        vi.resetAllMocks();
        // Default mock implementation
        (api.useParties as import('vitest').Mock).mockReturnValue({
            data: mockParties,
            isLoading: false,
            error: null,
            refetch: vi.fn(),
            isFetching: false
        });
    });

    it('renders the search input and table headers', () => {
        render(<PartySelector onSelect={mockOnSelect} />);

        // Search input
        expect(screen.getByPlaceholderText(/search/i)).toBeInTheDocument();

        // Table headers (checking for common headers)
        expect(screen.getByRole('columnheader', { name: /type/i })).toBeInTheDocument();
        expect(screen.getByRole('columnheader', { name: /name/i })).toBeInTheDocument();
        expect(screen.getByRole('columnheader', { name: /identifier/i })).toBeInTheDocument();
    });

    it('displays a list of parties in the table', () => {
        render(<PartySelector onSelect={mockOnSelect} />);

        // Check for party names in the table
        expect(screen.getByText('John Doe')).toBeInTheDocument();
        expect(screen.getByText('Acme Corp')).toBeInTheDocument();

        // Check for types
        expect(screen.getByText('Individual')).toBeInTheDocument();
        expect(screen.getByText('Organization')).toBeInTheDocument();
    });

    it('filters parties when searching', async () => {
        const user = userEvent.setup();
        render(<PartySelector onSelect={mockOnSelect} />);

        const searchInput = screen.getByPlaceholderText(/search/i);

        // Mock the hook to return filtered results when search params change
        // Note: In a real integration, the hook handles this. Here we just verify the input works.
        // We can verify calls to the hook if we want strict unit testing.
        await user.type(searchInput, 'John');

        // Check if useParties was called with filter
        expect(api.useParties).toHaveBeenLastCalledWith(
            expect.objectContaining({ givenName: 'John', tradingName: 'John' })
        );
    });

    it('calls onSelect when a row is clicked', async () => {
        const user = userEvent.setup();
        render(<PartySelector onSelect={mockOnSelect} />);

        // Click on "Acme Corp"
        // We assume the row or a select button is clickable. 
        // Let's assume clicking the name or the row works.
        const row = screen.getByText('Acme Corp').closest('tr');
        expect(row).toBeInTheDocument();

        await user.click(screen.getByText('Acme Corp'));

        expect(mockOnSelect).toHaveBeenCalledTimes(1);
        expect(mockOnSelect).toHaveBeenCalledWith(
            expect.objectContaining({ tradingName: 'Acme Corp' })
        );
    });



    it('shows empty state when no parties found', () => {
        (api.useParties as import('vitest').Mock).mockReturnValue({
            data: [],
            isLoading: false,
            error: null,
            refetch: vi.fn(),
            isFetching: false
        });

        render(<PartySelector onSelect={mockOnSelect} />);
        expect(screen.getByText(/no parties found/i)).toBeInTheDocument();
    });
});
