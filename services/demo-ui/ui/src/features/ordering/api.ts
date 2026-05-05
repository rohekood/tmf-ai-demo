import { useMutation, useQuery } from '@tanstack/react-query';
import { apiClient } from '../../api/client';
import type {
    QualificationRequest,
    QualificationSession,
    CartItemRequest,
    Cart,
    CheckoutRequest,
    SagaStatusResponse
} from './types';

// UC-01: Qualification API
export const checkQualification = async (request: QualificationRequest): Promise<QualificationSession> => {
    const { data } = await apiClient.post<QualificationSession>('/api/qualification/check', request);
    return data;
};

export const getQualificationSession = async (sessionId: string): Promise<QualificationSession> => {
    const { data } = await apiClient.get<QualificationSession>(`/api/qualification/session/${sessionId}`);
    return data;
};

// UC-02: Cart API
export const addCartItem = async (request: CartItemRequest): Promise<void> => {
    await apiClient.post('/api/cart/items', request);
};

export const getCart = async (cartId: string): Promise<Cart> => {
    const { data } = await apiClient.get<Cart>(`/api/cart/${cartId}`);
    return data;
};

export const removeCartItem = async (cartId: string, itemId: string): Promise<void> => {
    await apiClient.delete(`/api/cart/${cartId}/items/${itemId}`);
};

// UC-03: Checkout/Saga API
export const checkout = async (request: CheckoutRequest): Promise<{ sagaId: string }> => {
    const { data } = await apiClient.post<{ sagaId: string }>('/api/orders/checkout', request);
    return data;
};

export const getSagaStatus = async (sagaId: string): Promise<SagaStatusResponse> => {
    const { data } = await apiClient.get<SagaStatusResponse>(`/api/orders/saga/${sagaId}`);
    return data;
};

// --- Hooks ---

export const useCheckQualification = () => {
    return useMutation({
        mutationFn: checkQualification,
    });
};

export const useQualificationSession = (sessionId?: string) => {
    return useQuery({
        queryKey: ['qualification', sessionId],
        queryFn: () => getQualificationSession(sessionId!),
        enabled: !!sessionId,
        retry: false, // Don't retry on 422 expired
    });
};

export const useCart = (cartId?: string) => {
    return useQuery({
        queryKey: ['cart', cartId],
        queryFn: () => getCart(cartId!),
        enabled: !!cartId,
    });
};

export const useAddCartItem = () => {
    return useMutation({
        mutationFn: addCartItem,
    });
};

export const useRemoveCartItem = () => {
    return useMutation({
        mutationFn: ({ cartId, itemId }: { cartId: string; itemId: string }) => removeCartItem(cartId, itemId),
    });
};

export const useCheckout = () => {
    return useMutation({
        mutationFn: checkout,
    });
};

export const useSagaStatus = (sagaId?: string) => {
    return useQuery({
        queryKey: ['sagaStatus', sagaId],
        queryFn: () => getSagaStatus(sagaId!),
        enabled: !!sagaId,
        refetchInterval: (query) => {
            const status = query.state.data?.status;
            return status === 'COMPLETED' || status === 'FAILED' ? false : 3000;
        },
    });
};
