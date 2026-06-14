
import { render, screen, fireEvent } from '@testing-library/react';
import { MessageFeed } from './MessageFeed';
import { describe, it, expect, vi } from 'vitest';
import type { DebugMessage } from '../types';

describe('MessageFeed', () => {
    const mockMessages: DebugMessage[] = [
        {
            id: 'msg-1',
            type: 'command',
            topic: 'cmd.party.create',
            service: 'bff',
            timestamp: new Date().toISOString(),
            payload: {},
        },
        {
            id: 'msg-2',
            type: 'event',
            topic: 'evt.party.created',
            service: 'party-management',
            timestamp: new Date().toISOString(),
            payload: {},
        },
    ];
    const mockOnSelect = vi.fn();

    it('renders empty state', () => {
        render(<MessageFeed messages={[]} selectedId={null} onSelect={mockOnSelect} />);
        expect(screen.getByText('No messages captured yet.')).toBeInTheDocument();
    });

    it('renders list of messages', () => {
        render(<MessageFeed messages={mockMessages} selectedId={null} onSelect={mockOnSelect} />);

        expect(screen.getByText('cmd.party.create')).toBeInTheDocument();
        expect(screen.getByText('evt.party.created')).toBeInTheDocument();
        expect(screen.getByText('bff')).toBeInTheDocument();
        expect(screen.getByText('party-management')).toBeInTheDocument();
    });

    it('highlights selected message', () => {
        const { container } = render(
            <MessageFeed messages={mockMessages} selectedId="msg-1" onSelect={mockOnSelect} />
        );

        // Check for 'selected' class
        const selectedItem = container.querySelector('.feed-item.selected');
        expect(selectedItem).toBeInTheDocument();
        expect(selectedItem).toHaveTextContent('cmd.party.create');
    });

    it('calls onSelect when message clicked', () => {
        render(<MessageFeed messages={mockMessages} selectedId={null} onSelect={mockOnSelect} />);

        fireEvent.click(screen.getByText('cmd.party.create'));
        expect(mockOnSelect).toHaveBeenCalledWith(mockMessages[0]);
    });

    it('renders query/unknown types and a truncated correlation id', () => {
        const messages: DebugMessage[] = [
            { id: 'q', type: 'query', topic: 'qry.get', service: 'svc', timestamp: new Date().toISOString(), payload: {}, correlationId: 'abcdef12345678' },
            { id: 'u', type: 'unknown', topic: 'unk', service: 'svc', timestamp: new Date().toISOString(), payload: {} },
        ];
        render(<MessageFeed messages={messages} selectedId={null} onSelect={mockOnSelect} />);

        expect(screen.getByText('qry.get')).toBeInTheDocument();
        expect(screen.getByText('unk')).toBeInTheDocument();
        // correlationId present -> last 8 chars shown
        expect(screen.getByText('CID: 12345678')).toBeInTheDocument();
    });
});
