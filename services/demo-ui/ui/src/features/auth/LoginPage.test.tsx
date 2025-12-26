import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { LoginPage } from './LoginPage';

const mockLoginWithRedirect = vi.fn();

vi.mock('@auth0/auth0-react', () => ({
    useAuth0: () => ({
        loginWithRedirect: mockLoginWithRedirect,
    }),
}));

describe('LoginPage', () => {
    it('renders correctly', () => {
        render(<LoginPage />);
        expect(screen.getByText('TMF Demo')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /log in/i })).toBeInTheDocument();
    });

    it('calls loginWithRedirect when login button is clicked', () => {
        render(<LoginPage />);
        const button = screen.getByRole('button', { name: /log in/i });
        fireEvent.click(button);
        expect(mockLoginWithRedirect).toHaveBeenCalledWith(expect.objectContaining({
            authorizationParams: { screen_hint: 'login' }
        }));
    });
});
