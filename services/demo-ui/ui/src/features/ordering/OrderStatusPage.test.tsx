import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import OrderStatusPage from './OrderStatusPage';
import * as api from './api';

vi.mock('./api', () => ({
    useSagaStatus: vi.fn(),
}));

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom');
    return {
        ...actual,
        useNavigate: () => mockNavigate,
    };
});

describe('OrderStatusPage', () => {
    it('renders loading state initially', () => {
        vi.mocked(api.useSagaStatus).mockReturnValue({
            data: undefined,
            isLoading: true,
        } as any);

        render(
            <MemoryRouter initialEntries={['/order/status/saga1']}>
                <Routes>
                    <Route path="/order/status/:sagaId" element={<OrderStatusPage />} />
                </Routes>
            </MemoryRouter>
        );

        expect(screen.getByText('Loading...')).toBeInTheDocument();
    });

    it('navigates to confirmation when completed', () => {
        vi.useFakeTimers();
        vi.mocked(api.useSagaStatus).mockReturnValue({
            data: { status: 'COMPLETED', orderId: 'order123' },
            isLoading: false,
        } as any);

        render(
            <MemoryRouter initialEntries={['/order/status/saga1']}>
                <Routes>
                    <Route path="/order/status/:sagaId" element={<OrderStatusPage />} />
                </Routes>
            </MemoryRouter>
        );

        vi.runAllTimers();
        expect(mockNavigate).toHaveBeenCalledWith('/order/confirmation/order123');
        vi.useRealTimers();
    });

    it('renders failure state when saga fails', () => {
        vi.mocked(api.useSagaStatus).mockReturnValue({
            data: { status: 'FAILED', errorReason: 'Payment rejected' },
            isLoading: false,
        } as any);

        render(
            <MemoryRouter initialEntries={['/order/status/saga1']}>
                <Routes>
                    <Route path="/order/status/:sagaId" element={<OrderStatusPage />} />
                </Routes>
            </MemoryRouter>
        );

        expect(screen.getByText('Order processing failed')).toBeInTheDocument();
        expect(screen.getByText('Payment rejected')).toBeInTheDocument();
    });
});
