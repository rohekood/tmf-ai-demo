import { useNavigate } from 'react-router-dom';
import { useQueryClient } from '@tanstack/react-query';
import { useCart, useRemoveCartItem } from './api';
import { PageLoader } from '../../design-system/components/common/PageLoader';

const CART_ID_KEY = 'cartId';

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
    if (error) return <div className="p-8 text-red-600">Failed to load cart.</div>;

    const isEmpty = !cart || !cart.items || cart.items.length === 0;

    return (
        <div className="p-8 max-w-4xl mx-auto space-y-8">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold">Shopping Cart</h1>
                    <p className="text-gray-500 mt-2">Review your selected services.</p>
                </div>
            </div>

            {isEmpty ? (
                <div className="bg-white p-12 text-center rounded-lg shadow border border-gray-100">
                    <p className="text-gray-500 text-lg">Your cart is empty.</p>
                    <button
                        onClick={() => navigate('/order/qualify')}
                        className="mt-4 text-blue-600 hover:underline"
                    >
                        Browse Services
                    </button>
                </div>
            ) : (
                <div className="bg-white rounded-lg shadow border border-gray-100 overflow-hidden">
                    <ul className="divide-y divide-gray-200">
                        {cart.items.map((item) => (
                            <li key={item.id} className="p-6 flex items-center justify-between hover:bg-gray-50">
                                <div>
                                    <h3 className="text-lg font-medium text-gray-900">
                                        {item.name || `Offering: ${item.offeringId}`}
                                    </h3>
                                    <p className="text-sm text-gray-500">Quantity: {item.quantity}</p>
                                </div>
                                <div className="flex items-center space-x-6">
                                    <p className="text-lg font-semibold text-gray-900">
                                        {item.price} {item.currency}
                                    </p>
                                    <button
                                        onClick={() => handleRemove(item.id)}
                                        disabled={isRemoving}
                                        className="text-red-600 hover:text-red-900 text-sm font-medium disabled:opacity-50"
                                    >
                                        Remove
                                    </button>
                                </div>
                            </li>
                        ))}
                    </ul>
                    <div className="bg-gray-50 p-6 border-t border-gray-200 flex items-center justify-between">
                        <div>
                            <p className="text-sm text-gray-500">Total</p>
                            <p className="text-2xl font-bold text-gray-900">
                                {cart.totalPrice} {cart.currency}
                            </p>
                        </div>
                        <button
                            onClick={() => navigate('/order/checkout')}
                            className="bg-blue-600 text-white px-8 py-3 rounded-md shadow hover:bg-blue-700 font-medium"
                        >
                            Proceed to Checkout
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
}
