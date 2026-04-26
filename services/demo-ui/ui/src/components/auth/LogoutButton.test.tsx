import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { LogoutButton } from './LogoutButton';

const mockLogout = vi.fn();

vi.mock('../../auth/context', () => ({
    useAuth: () => ({
        logout: mockLogout,
    }),
}));

describe('LogoutButton', () => {
    it('renders logout button', () => {
        render(<LogoutButton />);
        expect(screen.getByRole('button', { name: /log out/i })).toBeInTheDocument();
    });

    it('calls logout on click', () => {
        render(<LogoutButton />);
        fireEvent.click(screen.getByRole('button', { name: /log out/i }));
        expect(mockLogout).toHaveBeenCalled();
    });
});
