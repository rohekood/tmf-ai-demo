import { useMutation, useQuery } from '@tanstack/react-query';
import { apiClient } from '../../api/client';

export type ProvisionStatus = 'ready' | 'needs_party' | 'needs_customer';

export interface ResolveCustomerResponse {
    status: ProvisionStatus;
    partyId?: string;
    customer?: { id: string;[key: string]: unknown };
}

export interface ProvisionRequest {
    givenName: string;
    familyName: string;
    phone?: string;
    street?: string;
    city?: string;
    postcode?: string;
    country?: string;
}

export const resolveCustomer = async (): Promise<ResolveCustomerResponse> => {
    const { data } = await apiClient.get<ResolveCustomerResponse>('/api/me/customer');
    return data;
};

export const provisionCustomer = async (request: ProvisionRequest): Promise<ResolveCustomerResponse> => {
    const { data } = await apiClient.post<ResolveCustomerResponse>('/api/me/provision', request);
    return data;
};

export const meCustomerQueryKey = ['me', 'customer'] as const;

/**
 * Resolves whether the logged-in user is backed by a customer. Only enabled once
 * the caller is authenticated (otherwise the request would 401).
 */
export const useResolveCustomer = (enabled: boolean) => {
    return useQuery({
        queryKey: meCustomerQueryKey,
        queryFn: resolveCustomer,
        enabled,
        staleTime: 5 * 60 * 1000,
    });
};

export const useProvisionCustomer = () => {
    return useMutation<ResolveCustomerResponse, Error, ProvisionRequest>({
        mutationFn: provisionCustomer,
    });
};
