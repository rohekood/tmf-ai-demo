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

// Mock useCart
vi.mock("../../features/ordering/api", () => ({
    useCart: vi.fn(() => ({ data: undefined })),
}));

import * as orderingApi from "../../features/ordering/api";

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
            vi.mocked(orderingApi.useCart).mockReturnValue({ data: undefined } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

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
            vi.mocked(orderingApi.useCart).mockReturnValue({ data: undefined } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

            render(
                <MemoryRouter>
                    <Sidebar collapsed={false} mobileOpen={false} onToggleCollapse={() => { }} />
                </MemoryRouter>
            );

            const link = screen.getByRole('link', { name: /Check Availability/i });
            expect(link).toHaveAttribute('href', '/order/qualify');
        });

        it('renders "Shopping Cart" link pointing to /order/cart', () => {
            vi.mocked(orderingApi.useCart).mockReturnValue({ data: undefined } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

            render(
                <MemoryRouter>
                    <Sidebar collapsed={false} mobileOpen={false} onToggleCollapse={() => { }} />
                </MemoryRouter>
            );

            const link = screen.getByRole('link', { name: /Shopping Cart/i });
            expect(link).toHaveAttribute('href', '/order/cart');
        });

        it('shows cart item count badge when cart has items', () => {
            vi.mocked(orderingApi.useCart).mockReturnValue({
                data: { id: 'default-cart', items: [{ id: 'i1' }, { id: 'i2' }] }
            } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

            render(
                <MemoryRouter>
                    <Sidebar collapsed={false} mobileOpen={false} onToggleCollapse={() => { }} />
                </MemoryRouter>
            );

            expect(screen.getByLabelText('2 items')).toBeInTheDocument();
            expect(screen.getByLabelText('2 items')).toHaveTextContent('2');
        });

        it('does not show cart badge when cart is empty', () => {
            vi.mocked(orderingApi.useCart).mockReturnValue({
                data: { id: 'default-cart', items: [] }
            } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

            render(
                <MemoryRouter>
                    <Sidebar collapsed={false} mobileOpen={false} onToggleCollapse={() => { }} />
                </MemoryRouter>
            );

            expect(screen.queryByLabelText(/items/i)).not.toBeInTheDocument();
        });

        it('does not show cart badge when cart data is unavailable', () => {
            vi.mocked(orderingApi.useCart).mockReturnValue({ data: undefined } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

            render(
                <MemoryRouter>
                    <Sidebar collapsed={false} mobileOpen={false} onToggleCollapse={() => { }} />
                </MemoryRouter>
            );

            expect(screen.queryByLabelText(/items/i)).not.toBeInTheDocument();
        });
    });
});
