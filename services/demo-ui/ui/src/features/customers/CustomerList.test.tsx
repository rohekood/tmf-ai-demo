
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { CustomerList } from './CustomerList';
import { apiClient } from '../../api/client';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { vi, describe, it, expect, beforeEach } from 'vitest';

// Mock apiClient
vi.mock('../../api/client', () => ({
    apiClient: {
        get: vi.fn(),
        post: vi.fn(),
    },
}));

const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            retry: false,
        },
    },
});

const renderWithClient = (ui: React.ReactNode) => {
    return render(
        <QueryClientProvider client={queryClient}>
            {ui}
        </QueryClientProvider>
    );
};

describe('CustomerList', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        queryClient.clear();
    });

    it('renders loading state initially', () => {
        (apiClient.get as any).mockImplementation(() => new Promise(() => { })); // Never resolves
        renderWithClient(<CustomerList />);
        expect(screen.getByText('Loading customers...')).toBeInTheDocument();
    });

    it('renders customers when API call succeeds', async () => {
        const mockCustomers = [
            { id: '1', name: 'John Doe', status: 'Active' },
            { id: '2', name: 'Jane Smith', status: 'Suspended' },
        ];
        (apiClient.get as any).mockResolvedValue({ data: mockCustomers });

        renderWithClient(<CustomerList />);

        await waitFor(() => {
            expect(screen.getByText('John Doe')).toBeInTheDocument();
            expect(screen.getByText('Jane Smith')).toBeInTheDocument();
        });
        expect(screen.getByText('Active')).toBeInTheDocument();
        expect(screen.getByText('Suspended')).toBeInTheDocument();
    });

    it('renders empty state when no customers returned', async () => {
        (apiClient.get as any).mockResolvedValue({ data: [] });

        renderWithClient(<CustomerList />);

        await waitFor(() => {
            expect(screen.getByText('No customers found.')).toBeInTheDocument();
        });
    });

    it('handles API error', async () => {
        (apiClient.get as any).mockRejectedValue(new Error('API Failure'));

        renderWithClient(<CustomerList />);

        await waitFor(() => {
            expect(screen.getByText('Error: API Failure')).toBeInTheDocument();
        });
    });

    it('triggers search when typing', async () => {
        (apiClient.get as any).mockResolvedValue({ data: [] });
        renderWithClient(<CustomerList />);

        const input = screen.getByPlaceholderText('Search by name...');
        fireEvent.change(input, { target: { value: 'Alice' } });

        await waitFor(() => {
            expect(apiClient.get).toHaveBeenCalledWith('/api/customers', { params: { name: 'Alice' } });
        });
    });

    it('creates demo customer', async () => {
        (apiClient.get as any).mockResolvedValue({ data: [] });
        (apiClient.post as any).mockResolvedValue({ data: { id: 'new-1', name: 'Demo Customer 123' } });

        renderWithClient(<CustomerList />);

        const button = screen.getByText('+ New Demo Customer');
        fireEvent.click(button);

        await waitFor(() => {
            expect(apiClient.post).toHaveBeenCalled();
        });
        // Should invalidate query and refetch
        expect(apiClient.get).toHaveBeenCalledTimes(2); // Initial + Refetch
    });
});
