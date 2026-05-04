export interface QualificationRequest {
    address: {
        street: string;
        city: string;
        postcode: string;
        country: string;
    };
    categoryIds?: string[];
}

export interface QualificationSession {
    id: string;
    state: string;
    address: {
        street: string;
        city: string;
        postcode: string;
        country: string;
    };
    qualifiedOfferings: Array<{
        offeringId: string;
        name: string;
        description?: string;
        price: number;
        currency: string;
    }>;
}

export interface CartItemRequest {
    cartId?: string;
    offeringId: string;
    qualificationSessionId?: string;
    quantity: number;
}

export interface AddCartItemResponse {
    cartId: string;
    items: CartItem[];
    totalPrice: number;
    currency: string;
}

export interface CartItem {
    id: string;
    offeringId: string;
    quantity: number;
    price: number;
    currency: string;
    name?: string; // Optional if BFF resolves it
}

export interface Cart {
    id: string;
    state: string;
    items: CartItem[];
    totalPrice: number;
    currency: string;
}

export interface CheckoutRequest {
    cartId: string;
    customerId: string;
    paymentDetails: {
        method: string;
        token: string;
    };
}

export interface SagaStatusResponse {
    id: string;
    cartId: string;
    status: 'STARTED' | 'INVENTORY_RESERVED' | 'INVENTORY_FAILED' | 'PAYMENT_AUTHORIZED' | 'PAYMENT_FAILED' | 'COMPLETED' | 'FAILED';
    orderId?: string;
    errorReason?: string;
}
