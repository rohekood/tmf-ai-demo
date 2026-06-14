import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { EmptyState } from './EmptyState';

describe('EmptyState', () => {
    it('renders the title with a status role', () => {
        render(<EmptyState title="No parties found." />);
        expect(screen.getByRole('status')).toHaveTextContent('No parties found.');
    });

    it('renders description and action when provided', () => {
        render(
            <EmptyState
                title="Your cart is empty."
                description="Browse the catalog to add services."
                action={<button>Browse Services</button>}
            />
        );
        expect(screen.getByText('Browse the catalog to add services.')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Browse Services' })).toBeInTheDocument();
    });

    it('marks the icon as decorative (aria-hidden)', () => {
        const { container } = render(
            <EmptyState title="Empty" icon={<svg data-testid="icon" />} />
        );
        const iconWrap = container.querySelector('.empty-state-block__icon');
        expect(iconWrap).toHaveAttribute('aria-hidden', 'true');
    });

    it('omits description and action when not provided', () => {
        const { container } = render(<EmptyState title="Empty" />);
        expect(container.querySelector('.empty-state-block__description')).toBeNull();
        expect(container.querySelector('.empty-state-block__action')).toBeNull();
    });
});
