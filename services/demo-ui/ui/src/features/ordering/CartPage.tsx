import { useNavigate } from 'react-router-dom';
import { useQueryClient } from '@tanstack/react-query';
import { useCart, useRemoveCartItem } from './api';
import { PageLoader } from '../../design-system/components/common/PageLoader';
import { EmptyState } from '../../design-system/components/common/EmptyState';
import { ShoppingCart } from 'lucide-react';
import { CART_ID_KEY } from './storage';
import './ordering.css';

export default function CartPage() {
    const navigate = useNavigate();
    const queryClient = useQueryClient();
    const cartId = localStorage.getItem(CART_ID_KEY) || undefined;

    const { data: cart, isLoading, error } = useCart(cartId);
    const { mutate: removeItem, isPending: isRemoving } = useRemoveCartItem();

    const handleRemove = (itemId: string) => {
        if (!cartId) return;
        removeItem(
            { cartId, itemId },
            {
                onSuccess: () => {
                    queryClient.invalidateQueries({ queryKey: ['cart', cartId] });
                }
            }
        );
    };

    if (isLoading) return <PageLoader />;
    if (error) return <div className="page"><div className="alert alert-danger">Failed to load cart.</div></div>;

    const isEmpty = !cart || !cart.items || cart.items.length === 0;

    return (
        <div className="page page--narrow">
            <div className="page-header">
                <div>
                    <h1 className="page-title">Shopping Cart</h1>
                    <p className="page-subtitle">Review your selected services.</p>
                </div>
            </div>

            {isEmpty ? (
                <EmptyState
                    icon={<ShoppingCart size={48} />}
                    title="Your cart is empty."
                    description="Check service availability for an address to add offerings to your cart."
                    action={
                        <button onClick={() => navigate('/order/qualify')} className="btn btn-primary">
                            Browse Services
                        </button>
                    }
                />
            ) : (
                <div className="card card--flush">
                    <ul className="data-list">
                        {cart.items.map((item) => (
                            <li key={item.id}>
                                <div>
                                    <h3 className="cart-item__name">
                                        {item.name || `Offering: ${item.offeringId}`}
                                    </h3>
                                    <p className="cart-item__qty">Quantity: {item.quantity}</p>
                                </div>
                                <div className="row" style={{ gap: '1.5rem' }}>
                                    <p className="cart-item__price">
                                        {item.price} {item.currency}
                                    </p>
                                    <button
                                        onClick={() => handleRemove(item.id)}
                                        disabled={isRemoving}
                                        className="btn-link"
                                        style={{ color: 'var(--danger)' }}
                                    >
                                        Remove
                                    </button>
                                </div>
                            </li>
                        ))}
                    </ul>
                    <div className="card-footer row-between">
                        <div>
                            <p className="cart-total__label">Total</p>
                            <p className="cart-total__amount">
                                {cart.totalPrice} {cart.currency}
                            </p>
                        </div>
                        <button onClick={() => navigate('/order/checkout')} className="btn btn-primary btn--lg">
                            Proceed to Checkout
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
}
