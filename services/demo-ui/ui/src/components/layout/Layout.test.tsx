import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { Layout } from './Layout';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

vi.mock('../../auth/context', () => ({
    useAuth: () => ({
        isAuthenticated: true,
        isLoading: false,
        user: { name: 'Test User', email: 'test@example.com' },
        logout: vi.fn(),
        loginWithRedirect: vi.fn(),
    }),
}));

vi.mock('../../features/ordering/api', () => ({
    useCart: vi.fn(() => ({ data: undefined })),
}));

const createTestQueryClient = () => new QueryClient({
    defaultOptions: { queries: { retry: false } },
});

describe('Layout', () => {
    it('renders the layout with sidebar and headers', () => {
        render(
            <QueryClientProvider client={createTestQueryClient()}>
                <MemoryRouter>
                    <Layout />
                </MemoryRouter>
            </QueryClientProvider>
        );

        // Check header
        expect(screen.getByRole('banner')).toBeInTheDocument();
        expect(screen.getByRole('heading', { name: 'TMF Demo Dashboard' })).toBeInTheDocument();
        // Assuming the subtitle isn't a heading, maybe text is fine or role="note"
        expect(screen.getByText('Managed via Golang BFF & RabbitMQ')).toBeInTheDocument();

        // Check Sidebar presence
        expect(screen.getByRole('navigation')).toBeInTheDocument();
    });
});
