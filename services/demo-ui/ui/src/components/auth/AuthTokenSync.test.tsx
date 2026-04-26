import { render, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { AuthTokenSync } from './AuthTokenSync';
import * as clientApi from '../../api/client';

const mockGetAccessTokenSilently = vi.fn();
const mockUseAuth = vi.fn();

vi.mock('../../auth/context', () => ({
    useAuth: () => mockUseAuth(),
}));

// Spy on setAuthToken
const setAuthTokenSpy = vi.spyOn(clientApi, 'setAuthToken');

describe('AuthTokenSync', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('sets auth token when authenticated', async () => {
        mockUseAuth.mockReturnValue({
            isAuthenticated: true,
            getAccessTokenSilently: mockGetAccessTokenSilently,
        });
        mockGetAccessTokenSilently.mockResolvedValue('fake-token');

        render(<AuthTokenSync />);

        await waitFor(() => {
            expect(mockGetAccessTokenSilently).toHaveBeenCalled();
            expect(setAuthTokenSpy).toHaveBeenCalledWith('fake-token');
        });
    });

    it('clears auth token when not authenticated', async () => {
        mockUseAuth.mockReturnValue({
            isAuthenticated: false,
            getAccessTokenSilently: mockGetAccessTokenSilently,
        });

        render(<AuthTokenSync />);

        await waitFor(() => {
            expect(setAuthTokenSpy).toHaveBeenCalledWith(null);
        });
    });
});
