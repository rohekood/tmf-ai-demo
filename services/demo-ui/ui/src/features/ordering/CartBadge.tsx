import { useCart } from './api';
import { CART_ID_KEY } from './storage';
import './ordering.css';

export function CartBadge() {
    const cartId = localStorage.getItem(CART_ID_KEY) || undefined;
    const { data: cart } = useCart(cartId);

    const itemCount = cart?.items?.length ?? 0;

    if (itemCount === 0) return null;

    return (
        <span
            className="cart-badge"
            aria-label={`${itemCount} item${itemCount !== 1 ? 's' : ''} in cart`}
        >
            {itemCount}
        </span>
    );
}
