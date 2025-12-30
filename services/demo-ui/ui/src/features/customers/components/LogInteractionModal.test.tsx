import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import LogInteractionModal from './LogInteractionModal';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import * as api from '../api';

// Mock the API hook
vi.mock('../api', () => ({
    useLogInteraction: vi.fn()
}));

describe('LogInteractionModal', () => {
    const mockMutateAsync = vi.fn();
    const mockOnClose = vi.fn();
    const customerId = 'cust-123';

    beforeEach(() => {
        vi.mocked(api.useLogInteraction).mockReturnValue({
            mutateAsync: mockMutateAsync,
            isPending: false
        } as any);
    });

    it('renders correctly', () => {
        render(<LogInteractionModal customerId={customerId} onClose={mockOnClose} />);
        expect(screen.getByRole('heading', { level: 3, name: 'Log Interaction' })).toBeInTheDocument();
        expect(screen.getByLabelText('Type')).toBeInTheDocument();
        expect(screen.getByLabelText('Description')).toBeInTheDocument();
    });

    it('submits form with correct data', async () => {
        render(<LogInteractionModal customerId={customerId} onClose={mockOnClose} />);

        fireEvent.change(screen.getByLabelText('Type'), { target: { value: 'Email' } });
        fireEvent.change(screen.getByLabelText('Channel'), { target: { value: 'Web' } });
        fireEvent.change(screen.getByLabelText('Description'), { target: { value: 'Test interaction' } });

        fireEvent.click(screen.getByRole('button', { name: 'Log Interaction' }));

        await waitFor(() => {
            expect(mockMutateAsync).toHaveBeenCalledWith(expect.objectContaining({
                customerId: customerId,
                type: 'Email',
                channel: 'Web',
                description: 'Test interaction',
                id: expect.any(String), // UUID generated
                interactionDate: expect.any(String)
            }));
            expect(mockOnClose).toHaveBeenCalled();
        });
    });

    it('closes on cancel', () => {
        render(<LogInteractionModal customerId={customerId} onClose={mockOnClose} />);
        fireEvent.click(screen.getByText('Cancel'));
        expect(mockOnClose).toHaveBeenCalled();
    });
});
