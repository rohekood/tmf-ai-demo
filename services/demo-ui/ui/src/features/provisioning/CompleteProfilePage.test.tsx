import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import CompleteProfilePage from './CompleteProfilePage';
import * as api from './api';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom');
    return { ...actual, useNavigate: () => mockNavigate };
});

vi.mock('../../auth/context', () => ({
    useAuth: () => ({ user: { email: 'jane@example.com', name: 'Jane Doe' } }),
}));

const showToast = vi.fn();
vi.mock('../../design-system/components/common/Toast', () => ({
    useNotification: () => ({ showToast }),
}));

function renderPage() {
    const qc = new QueryClient();
    return render(
        <QueryClientProvider client={qc}>
            <MemoryRouter>
                <CompleteProfilePage />
            </MemoryRouter>
        </QueryClientProvider>
    );
}

describe('CompleteProfilePage', () => {
    it('prefills email (read-only) and splits the name', () => {
        vi.spyOn(api, 'useProvisionCustomer').mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
        } as unknown as ReturnType<typeof api.useProvisionCustomer>);

        renderPage();

        const email = screen.getByLabelText(/Email/i) as HTMLInputElement;
        expect(email.value).toBe('jane@example.com');
        expect(email).toHaveAttribute('readonly');
        expect((screen.getByLabelText(/First name/i) as HTMLInputElement).value).toBe('Jane');
        expect((screen.getByLabelText(/Last name/i) as HTMLInputElement).value).toBe('Doe');
    });

    it('submits the profile and navigates on success', async () => {
        const mutate = vi.fn().mockImplementation((_req, opts) => opts.onSuccess());
        vi.spyOn(api, 'useProvisionCustomer').mockReturnValue({
            mutate,
            isPending: false,
        } as unknown as ReturnType<typeof api.useProvisionCustomer>);

        renderPage();

        fireEvent.click(screen.getByRole('button', { name: /Save and continue/i }));

        expect(mutate).toHaveBeenCalledWith(
            expect.objectContaining({ givenName: 'Jane', familyName: 'Doe' }),
            expect.anything()
        );
        await waitFor(() => expect(mockNavigate).toHaveBeenCalled());
    });
});
