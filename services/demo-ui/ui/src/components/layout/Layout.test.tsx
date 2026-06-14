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

vi.mock('../../features/ordering/CartBadge', () => ({
    CartBadge: () => null,
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

        // Check header. The banner shows app branding as plain text (not a
        // heading) so each page can own the document <h1>.
        expect(screen.getByRole('banner')).toBeInTheDocument();
        expect(screen.getByText('TMF Demo Dashboard')).toBeInTheDocument();
        expect(
            screen.queryByRole('heading', { name: 'TMF Demo Dashboard' })
        ).not.toBeInTheDocument();
        expect(screen.getByText('Managed via Golang BFF & RabbitMQ')).toBeInTheDocument();

        // Check Sidebar presence
        expect(screen.getByRole('navigation')).toBeInTheDocument();
    });
});
