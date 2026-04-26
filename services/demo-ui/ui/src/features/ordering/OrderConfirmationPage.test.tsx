import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import OrderConfirmationPage from './OrderConfirmationPage';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom');
    return {
        ...actual,
        useNavigate: () => mockNavigate,
    };
});

describe('OrderConfirmationPage', () => {
    it('renders confirmation message and allows return to dashboard', () => {
        render(
            <MemoryRouter>
                <OrderConfirmationPage />
            </MemoryRouter>
        );

        expect(screen.getByText('Order Confirmed!')).toBeInTheDocument();
        
        const link = screen.getByRole('link', { name: /Return to Dashboard/i });
        expect(link).toBeInTheDocument();
        expect(link).toHaveAttribute('href', '/');
    });
});
