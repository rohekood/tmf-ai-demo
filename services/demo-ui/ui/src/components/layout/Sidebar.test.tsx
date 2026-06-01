import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { Sidebar } from './Sidebar';
import { MemoryRouter } from 'react-router-dom';

import { vi } from 'vitest';

// Mock useAuth
vi.mock("../../auth/context", () => ({
    useAuth: () => ({
        isAuthenticated: true,
        user: { name: 'Test User', email: 'test@example.com', picture: 'https://example.com/avatar.jpg' },
        logout: vi.fn(),
        loginWithRedirect: vi.fn(),
        isLoading: false
    })
}));

vi.mock('../../features/ordering/CartBadge', () => ({
    CartBadge: () => null,
}));

describe('Sidebar', () => {
    it('renders navigation links', () => {
        const props = {
            collapsed: false,
            mobileOpen: false,
            onToggleCollapse: () => { },
        };

        render(
            <MemoryRouter>
                <Sidebar {...props} />
            </MemoryRouter>
        );

        expect(screen.getByRole('link', { name: 'Parties' })).toBeInTheDocument();
        expect(screen.getByRole('link', { name: 'Customers' })).toBeInTheDocument();
        expect(screen.getByRole('link', { name: 'Debug Console' })).toBeInTheDocument();
        expect(screen.getByRole('status')).toHaveTextContent('Connected');
    });

    it('renders user avatar but hides info when collapsed', () => {
        render(
            <MemoryRouter>
                <Sidebar collapsed={true} mobileOpen={false} onToggleCollapse={() => { }} />
            </MemoryRouter>
        );

        // Avatar should be visible
        expect(screen.getByRole('img', { name: 'Test User' })).toBeInTheDocument();

        // User details group and Logout button should be hidden (removed from DOM)
        expect(screen.queryByRole('group', { name: /user details/i })).not.toBeInTheDocument();
        expect(screen.queryByRole('button', { name: /log out/i })).not.toBeInTheDocument();
    });

    describe('Ordering section', () => {
        it('renders "Ordering" section heading and both nav links', () => {
            render(
                <MemoryRouter>
                    <Sidebar collapsed={false} mobileOpen={false} onToggleCollapse={() => { }} />
                </MemoryRouter>
            );

            expect(screen.getByText('Ordering')).toBeInTheDocument();
            expect(screen.getByRole('link', { name: /Check Availability/i })).toBeInTheDocument();
            expect(screen.getByRole('link', { name: /Shopping Cart/i })).toBeInTheDocument();
        });

        it('renders "Check Availability" link pointing to /order/qualify', () => {
            render(
                <MemoryRouter>
                    <Sidebar collapsed={false} mobileOpen={false} onToggleCollapse={() => { }} />
                </MemoryRouter>
            );

            const link = screen.getByRole('link', { name: /Check Availability/i });
            expect(link).toHaveAttribute('href', '/order/qualify');
        });

        it('renders "Shopping Cart" link pointing to /order/cart', () => {
            render(
                <MemoryRouter>
                    <Sidebar collapsed={false} mobileOpen={false} onToggleCollapse={() => { }} />
                </MemoryRouter>
            );

            const link = screen.getByRole('link', { name: /Shopping Cart/i });
            expect(link).toHaveAttribute('href', '/order/cart');
        });
    });
});
