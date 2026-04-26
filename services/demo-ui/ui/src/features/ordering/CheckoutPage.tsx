import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useCart, useCheckout } from './api';
import { PageLoader } from '../../design-system/components/common/PageLoader';

export default function CheckoutPage() {
    const navigate = useNavigate();
    const cartId = 'default-cart';
    
    const { data: cart, isLoading } = useCart(cartId);
    const { mutate: submitCheckout, isPending } = useCheckout();

    const [paymentMethod, setPaymentMethod] = useState('credit_card');

    const handleCheckout = (e: React.FormEvent) => {
        e.preventDefault();
        submitCheckout(
            {
                cartId,
                customerId: 'demo-customer-id', // Would be from auth context
                paymentDetails: {
                    method: paymentMethod,
                    token: 'tok_visa_demo'
                }
            },
            {
                onSuccess: (data) => {
                    navigate(`/order/status/${data.sagaId}`);
                },
                onError: (err) => {
                    alert('Checkout failed: ' + err.message);
                }
            }
        );
    };

    if (isLoading) return <PageLoader />;

    return (
        <div className="p-8 max-w-4xl mx-auto space-y-8">
            <h1 className="text-3xl font-bold">Checkout</h1>
            
            <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                <div className="md:col-span-2 space-y-6">
                    <div className="bg-white p-6 rounded-lg shadow border border-gray-100">
                        <h2 className="text-xl font-semibold mb-4">Payment Method</h2>
                        <form id="checkout-form" onSubmit={handleCheckout} className="space-y-4">
                            <div className="space-y-2">
                                <label className="flex items-center space-x-3">
                                    <input
                                        type="radio"
                                        name="payment"
                                        value="credit_card"
                                        checked={paymentMethod === 'credit_card'}
                                        onChange={(e) => setPaymentMethod(e.target.value)}
                                        className="h-4 w-4 text-blue-600 border-gray-300 focus:ring-blue-500"
                                    />
                                    <span className="text-gray-900 font-medium">Credit Card</span>
                                </label>
                                <label className="flex items-center space-x-3">
                                    <input
                                        type="radio"
                                        name="payment"
                                        value="paypal"
                                        checked={paymentMethod === 'paypal'}
                                        onChange={(e) => setPaymentMethod(e.target.value)}
                                        className="h-4 w-4 text-blue-600 border-gray-300 focus:ring-blue-500"
                                    />
                                    <span className="text-gray-900 font-medium">PayPal</span>
                                </label>
                            </div>
                        </form>
                    </div>
                </div>

                <div className="md:col-span-1">
                    <div className="bg-gray-50 p-6 rounded-lg shadow border border-gray-200 sticky top-8">
                        <h2 className="text-lg font-semibold mb-4">Order Summary</h2>
                        {cart?.items.map(item => (
                            <div key={item.id} className="flex justify-between text-sm mb-2">
                                <span className="text-gray-600">{item.name || item.offeringId} (x{item.quantity})</span>
                                <span className="font-medium">{item.price} {item.currency}</span>
                            </div>
                        ))}
                        <div className="border-t border-gray-200 mt-4 pt-4 flex justify-between">
                            <span className="font-bold">Total</span>
                            <span className="font-bold text-lg">{cart?.totalPrice} {cart?.currency}</span>
                        </div>
                        <button
                            type="submit"
                            form="checkout-form"
                            disabled={isPending || !cart?.items.length}
                            className="w-full mt-6 bg-blue-600 text-white px-4 py-3 rounded-md shadow hover:bg-blue-700 font-medium disabled:opacity-50"
                        >
                            {isPending ? 'Processing...' : 'Place Order'}
                        </button>
                    </div>
                </div>
            </div>
            {isPending && <PageLoader />}
        </div>
    );
}
