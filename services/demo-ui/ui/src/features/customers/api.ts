import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../api/client';
import type { Customer, OnboardCustomerPayload, UpdateCustomerPayload, SearchCustomerParams, CustomerInteraction } from './types';

const CUSTOMERS_KEY = 'customers';

// Fetch all customers with optional search params
async function fetchCustomers(params?: SearchCustomerParams): Promise<Customer[]> {
    const searchParams = new URLSearchParams();
    if (params?.search) searchParams.set('search', params.search);
    if (params?.name) searchParams.set('name', params.name);
    if (params?.status) searchParams.set('status', params.status);
    if (params?.partyId) searchParams.set('partyId', params.partyId);

    const response = await apiClient.get<Customer[]>(`/api/customers?${searchParams.toString()}`);
    return response.data;
}

// Fetch single customer by ID
async function fetchCustomer(id: string): Promise<Customer> {
    const response = await apiClient.get<Customer>(`/api/customers/${id}`);
    return response.data;
}

// Onboard customer
async function onboardCustomer(payload: OnboardCustomerPayload): Promise<Customer> {
    const response = await apiClient.post<Customer>('/api/customers', payload);
    return response.data;
}

// Update customer
async function updateCustomer(payload: UpdateCustomerPayload): Promise<Customer> {
    const response = await apiClient.put<Customer>(`/api/customers/${payload.id}`, payload);
    return response.data;
}

// Delete customer
async function deleteCustomer(id: string): Promise<void> {
    await apiClient.delete(`/api/customers/${id}`);
}

// Log Interaction
async function logInteraction(payload: CustomerInteraction): Promise<void> {
    await apiClient.post(`/api/customers/${payload.customerId}/interactions`, payload);
}

// React Query Hooks

export function useCustomers(params?: SearchCustomerParams) {
    return useQuery({
        queryKey: [CUSTOMERS_KEY, params],
        queryFn: () => fetchCustomers(params),
    });
}

export function useCustomer(id: string | undefined) {
    return useQuery({
        queryKey: [CUSTOMERS_KEY, id],
        queryFn: () => fetchCustomer(id!),
        enabled: !!id,
    });
}

export function useOnboardCustomer() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: onboardCustomer,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: [CUSTOMERS_KEY] });
        },
    });
}

export function useUpdateCustomer() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: updateCustomer,
        onSuccess: (data) => {
            queryClient.invalidateQueries({ queryKey: [CUSTOMERS_KEY] });
            queryClient.setQueryData([CUSTOMERS_KEY, data.id], data);
        },
    });
}

export function useDeleteCustomer() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: deleteCustomer,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: [CUSTOMERS_KEY] });
        },
    });
}

export function useLogInteraction() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: logInteraction,
        onSuccess: (_, variables) => {
            queryClient.invalidateQueries({ queryKey: [CUSTOMERS_KEY, variables.customerId] });
        },
    });
}
