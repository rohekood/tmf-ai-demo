import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useCart, useCheckout } from './api';
import { PageLoader } from '../../design-system/components/common/PageLoader';
import './ordering.css';

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
        <div className="page page--narrow">
            <div className="page-header">
                <h1 className="page-title">Checkout</h1>
            </div>

            <div className="checkout-grid">
                <div className="card">
                    <h2 className="section-title" style={{ marginBottom: '1rem' }}>Payment Method</h2>
                    <form id="checkout-form" onSubmit={handleCheckout} className="stack-sm">
                        <label className="checkbox-row">
                            <input
                                type="radio"
                                name="payment"
                                value="credit_card"
                                checked={paymentMethod === 'credit_card'}
                                onChange={(e) => setPaymentMethod(e.target.value)}
                            />
                            <span>Credit Card</span>
                        </label>
                        <label className="checkbox-row">
                            <input
                                type="radio"
                                name="payment"
                                value="paypal"
                                checked={paymentMethod === 'paypal'}
                                onChange={(e) => setPaymentMethod(e.target.value)}
                            />
                            <span>PayPal</span>
                        </label>
                    </form>
                </div>

                <div className="card checkout-summary">
                    <h2 className="section-title" style={{ marginBottom: '1rem' }}>Order Summary</h2>
                    {cart?.items.map(item => (
                        <div key={item.id} className="summary-row">
                            <span>{item.name || item.offeringId} (x{item.quantity})</span>
                            <span style={{ color: 'var(--text)', fontWeight: 600 }}>{item.price} {item.currency}</span>
                        </div>
                    ))}
                    <div className="summary-total">
                        <span>Total</span>
                        <span>{cart?.totalPrice} {cart?.currency}</span>
                    </div>
                    <button
                        type="submit"
                        form="checkout-form"
                        disabled={isPending || !cart?.items.length}
                        className="btn btn-primary btn--block"
                        style={{ marginTop: '1.25rem' }}
                    >
                        {isPending ? 'Processing…' : 'Place Order'}
                    </button>
                </div>
            </div>
            {isPending && <PageLoader />}
        </div>
    );
}
