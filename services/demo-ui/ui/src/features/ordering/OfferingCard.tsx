import './ordering.css';
import type { QualifiedOffer } from './types';

interface OfferingCardProps {
    offering: QualifiedOffer;
    onAddToCart: (offeringId: string) => void;
    isAddingToCart: boolean;
}

export function OfferingCard({ offering, onAddToCart, isAddingToCart }: OfferingCardProps) {
    return (
        <div className="card card--hover offering-card">
            <div>
                <h3 className="offering-card__name">{offering.offeringName}</h3>
                <p className="offering-card__price">
                    {offering.price.amount}
                    <span className="offering-card__price-currency">{offering.price.currency}</span>
                </p>
                {offering.price.taxIncluded && (
                    <p className="offering-card__tax">Tax included</p>
                )}
                {offering.priceType === 'generic' && (
                    <p className="offering-card__price-note">Standard price — sign in for your pricing</p>
                )}
            </div>
            <button
                onClick={() => onAddToCart(offering.offeringId)}
                disabled={isAddingToCart || offering.eligibility !== 'QUALIFIED'}
                className="btn btn-success btn--block"
            >
                Add to Cart
            </button>
        </div>
    );
}
