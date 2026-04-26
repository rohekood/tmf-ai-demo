import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { LoginButton } from './LoginButton';

const mockLoginWithRedirect = vi.fn();

vi.mock('../../auth/context', () => ({
    useAuth: () => ({
        enabled: true,
        loginWithRedirect: mockLoginWithRedirect,
    }),
}));

describe('LoginButton', () => {
    it('renders login button', () => {
        render(<LoginButton />);
        expect(screen.getByRole('button', { name: /log in/i })).toBeInTheDocument();
    });

    it('calls loginWithRedirect on click', () => {
        render(<LoginButton />);
        fireEvent.click(screen.getByRole('button', { name: /log in/i }));
        expect(mockLoginWithRedirect).toHaveBeenCalledWith(expect.objectContaining({
            authorizationParams: { screen_hint: 'login' }
        }));
    });
});
