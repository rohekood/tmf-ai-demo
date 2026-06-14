import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import InteractionsList from './InteractionsList';
import type { CustomerInteraction } from '../types';

const items: CustomerInteraction[] = [
    {
        id: 'i1',
        customerId: 'c1',
        interactionDate: new Date(2026, 5, 14, 9, 30, 0).toISOString(),
        channel: 'phone',
        type: 'Support',
        description: 'Asked about billing',
        agentId: 'agent-7',
    },
];

describe('InteractionsList', () => {
    it('shows an empty message when there are no interactions', () => {
        render(<InteractionsList items={[]} />);
        expect(screen.getByText('No interactions logged')).toBeInTheDocument();
    });

    it('renders interactions with an Estonian-formatted date', () => {
        render(<InteractionsList items={items} />);
        expect(screen.getByText('by agent-7')).toBeInTheDocument();
        expect(screen.getByText('Support (phone)')).toBeInTheDocument();
        expect(screen.getByText('Asked about billing')).toBeInTheDocument();
        // dd.MM.yyyy HH:mm:ss (no comma)
        expect(screen.getByText(/^14\.06\.2026 \d{2}:\d{2}:\d{2}$/)).toBeInTheDocument();
    });
});
