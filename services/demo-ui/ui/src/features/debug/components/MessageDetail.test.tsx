
import { render, screen, fireEvent } from '@testing-library/react';
import { MessageDetail } from './MessageDetail';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { DebugMessage } from '../types';

describe('MessageDetail', () => {
    const mockMessage: DebugMessage = {
        id: 'msg-1',
        type: 'event',
        topic: 'tmf.events.customer.created',
        service: 'customer-management',
        timestamp: new Date().toISOString(),
        payload: { id: '123', name: 'Test' },
        correlationId: 'corr-123',
    };
    const mockOnClose = vi.fn();

    beforeEach(() => {
        // Mock clipboard
        Object.assign(navigator, {
            clipboard: {
                writeText: vi.fn(),
            },
        });
        vi.clearAllMocks();
    });

    it('renders empty state when no message selected', () => {
        render(<MessageDetail message={null} onClose={mockOnClose} />);
        expect(screen.getByText('Select a message to view details')).toBeInTheDocument();
    });

    it('renders message details', () => {
        render(<MessageDetail message={mockMessage} onClose={mockOnClose} />);

        expect(screen.getByText('event')).toBeInTheDocument();
        expect(screen.getByText('tmf.events.customer.created')).toBeInTheDocument();
        expect(screen.getByText('customer-management')).toBeInTheDocument();
        expect(screen.getByText('msg-1')).toBeInTheDocument();
        expect(screen.getByText('corr-123')).toBeInTheDocument();
        // Check pretty printed payload
        expect(screen.getByText(/"name": "Test"/)).toBeInTheDocument();
    });

    it('copies payload to clipboard', () => {
        render(<MessageDetail message={mockMessage} onClose={mockOnClose} />);

        const copyButton = screen.getByTitle('Copy Payload');
        fireEvent.click(copyButton);

        expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
            JSON.stringify(mockMessage.payload, null, 2)
        );
    });

    it('calls onClose when close button clicked', () => {
        render(<MessageDetail message={mockMessage} onClose={mockOnClose} />);

        const closeButton = screen.getByTitle('Close');
        fireEvent.click(closeButton);

        expect(mockOnClose).toHaveBeenCalled();
    });
});
