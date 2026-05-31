import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { OfferingCard } from './OfferingCard';
import type { QualifiedOffer } from './types';

const qualifiedOffer: QualifiedOffer = {
    offeringId: 'off-1',
    offeringName: 'Super Fiber 1Gbps',
    price: { amount: 49.99, currency: 'EUR', taxIncluded: false },
    eligibility: 'QUALIFIED',
};

describe('OfferingCard', () => {
    it('renders offering name, price, and currency', () => {
        render(
            <OfferingCard
                offering={qualifiedOffer}
                onAddToCart={vi.fn()}
                isAddingToCart={false}
            />
        );

        expect(screen.getByText('Super Fiber 1Gbps')).toBeInTheDocument();
        expect(screen.getByText(/49.99/)).toBeInTheDocument();
        expect(screen.getByText(/EUR/)).toBeInTheDocument();
    });

    it('does not show "Tax included" when taxIncluded is false', () => {
        render(
            <OfferingCard
                offering={qualifiedOffer}
                onAddToCart={vi.fn()}
                isAddingToCart={false}
            />
        );

        expect(screen.queryByText(/Tax included/i)).not.toBeInTheDocument();
    });

    it('shows "Tax included" when taxIncluded is true', () => {
        const withTax: QualifiedOffer = {
            ...qualifiedOffer,
            price: { ...qualifiedOffer.price, taxIncluded: true },
        };

        render(
            <OfferingCard
                offering={withTax}
                onAddToCart={vi.fn()}
                isAddingToCart={false}
            />
        );

        expect(screen.getByText(/Tax included/i)).toBeInTheDocument();
    });

    it('calls onAddToCart with the offering id when Add to Cart is clicked', async () => {
        const user = userEvent.setup();
        const mockAddToCart = vi.fn();

        render(
            <OfferingCard
                offering={qualifiedOffer}
                onAddToCart={mockAddToCart}
                isAddingToCart={false}
            />
        );

        await user.click(screen.getByRole('button', { name: /Add to Cart/i }));
        expect(mockAddToCart).toHaveBeenCalledWith('off-1');
    });

    it('disables Add to Cart button when isAddingToCart is true', () => {
        render(
            <OfferingCard
                offering={qualifiedOffer}
                onAddToCart={vi.fn()}
                isAddingToCart={true}
            />
        );

        expect(screen.getByRole('button', { name: /Add to Cart/i })).toBeDisabled();
    });

    it('disables Add to Cart button when eligibility is NOT_AVAILABLE', () => {
        const notAvailable: QualifiedOffer = {
            ...qualifiedOffer,
            eligibility: 'NOT_AVAILABLE',
        };

        render(
            <OfferingCard
                offering={notAvailable}
                onAddToCart={vi.fn()}
                isAddingToCart={false}
            />
        );

        expect(screen.getByRole('button', { name: /Add to Cart/i })).toBeDisabled();
    });

    it('enables Add to Cart button when eligibility is QUALIFIED and not adding', () => {
        render(
            <OfferingCard
                offering={qualifiedOffer}
                onAddToCart={vi.fn()}
                isAddingToCart={false}
            />
        );

        expect(screen.getByRole('button', { name: /Add to Cart/i })).not.toBeDisabled();
    });
});
