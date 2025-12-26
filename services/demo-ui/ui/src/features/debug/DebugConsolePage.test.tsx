import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import DebugConsolePage from './DebugConsolePage';
import { useDebugWebSocket } from './useDebugWebSocket';
import { type DebugMessage } from './types';

// Mock hook
vi.mock('./useDebugWebSocket');

const mockMessages: DebugMessage[] = [
    {
        id: 'msg1',
        timestamp: '2023-01-01T12:00:00Z',
        type: 'event',
        topic: 'tmf.party.created',
        payload: { id: 'p1', name: 'Party 1' },
        service: 'party'
    },
    {
        id: 'msg2',
        timestamp: '2023-01-01T12:01:00Z',
        type: 'command',
        topic: 'tmf.customer.onboard',
        payload: { name: 'Customer 1' },
        service: 'customer'
    }
];

describe('DebugConsolePage', () => {
    let clearMessagesMock: ReturnType<typeof vi.fn>;

    beforeEach(() => {
        vi.resetAllMocks();
        clearMessagesMock = vi.fn();
        vi.mocked(useDebugWebSocket).mockReturnValue({
            messages: mockMessages,
            isConnected: true,
            clearMessages: clearMessagesMock,
            error: null
        } as unknown as ReturnType<typeof useDebugWebSocket>);
    });

    it('renders connected status', () => {
        render(<DebugConsolePage />);
        expect(screen.getByRole('status')).toHaveTextContent('Live');
        expect(screen.getByText('2 events captured')).toBeInTheDocument();
    });

    it('renders message list', () => {
        render(<DebugConsolePage />);
        const items = screen.getAllByRole('listitem');
        expect(items).toHaveLength(2);
        expect(items[0]).toHaveTextContent('tmf.party.created');
        expect(items[1]).toHaveTextContent('tmf.customer.onboard');
    });

    it('filters by search text', async () => {
        const user = userEvent.setup();
        render(<DebugConsolePage />);
        const searchInput = screen.getByPlaceholderText('Search payload or topic...');
        await user.type(searchInput, 'party');

        const items = screen.getAllByRole('listitem');
        expect(items).toHaveLength(1);
        expect(items[0]).toHaveTextContent('tmf.party.created');
        expect(screen.queryByText('tmf.customer.onboard')).not.toBeInTheDocument();
    });

    it('filters by type', async () => {
        const user = userEvent.setup();
        render(<DebugConsolePage />);
        const eventChip = screen.getByRole('button', { name: 'event' });
        await user.click(eventChip);

        const items = screen.getAllByRole('listitem');
        expect(items).toHaveLength(1);
        expect(items[0]).toHaveTextContent('tmf.party.created');
        expect(screen.queryByText('tmf.customer.onboard')).not.toBeInTheDocument();
    });

    it('clears messages', async () => {
        const user = userEvent.setup();
        render(<DebugConsolePage />);
        const clearBtn = screen.getByRole('button', { name: /clear messages/i });
        await user.click(clearBtn);
        expect(clearMessagesMock).toHaveBeenCalled();
    });
});
