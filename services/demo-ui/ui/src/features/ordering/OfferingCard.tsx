import type { QualifiedOffer } from './types';

interface OfferingCardProps {
    offering: QualifiedOffer;
    onAddToCart: (offeringId: string) => void;
    isAddingToCart: boolean;
}

export function OfferingCard({ offering, onAddToCart, isAddingToCart }: OfferingCardProps) {
    return (
        <div className="bg-white p-6 rounded-lg shadow border border-gray-100 flex flex-col justify-between">
            <div>
                <h3 className="text-lg font-bold">{offering.offeringName}</h3>
                <p className="text-2xl font-semibold mt-4">
                    {offering.price.amount} {offering.price.currency}
                </p>
                {offering.price.taxIncluded && (
                    <p className="text-xs text-gray-400 mt-1">Tax included</p>
                )}
            </div>
            <button
                onClick={() => onAddToCart(offering.offeringId)}
                disabled={isAddingToCart || offering.eligibility !== 'QUALIFIED'}
                className="mt-6 bg-green-600 text-white px-4 py-2 rounded shadow hover:bg-green-700 disabled:opacity-50 w-full"
            >
                Add to Cart
            </button>
        </div>
    );
}
