import { useCart } from './api';

const CART_ID_KEY = 'cartId';

export function CartBadge() {
    const cartId = localStorage.getItem(CART_ID_KEY) || undefined;
    const { data: cart } = useCart(cartId);

    const itemCount = cart?.items?.length ?? 0;

    if (itemCount === 0) return null;

    return (
        <span
            className="ml-auto inline-flex items-center justify-center min-w-5 h-5 px-1 text-xs font-bold leading-none text-white bg-blue-600 rounded-full"
            aria-label={`${itemCount} item${itemCount !== 1 ? 's' : ''} in cart`}
        >
            {itemCount}
        </span>
    );
}
