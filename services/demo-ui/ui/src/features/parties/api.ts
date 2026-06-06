import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../api/client';
import type { PartyUnion, CreatePartyPayload, UpdatePartyPayload, SearchPartyParams } from './types';

const PARTIES_KEY = 'parties';

// Fetch all parties with optional search params
async function fetchParties(params?: SearchPartyParams): Promise<PartyUnion[]> {
    const searchParams = new URLSearchParams();
    if (params?.search) searchParams.set('search', params.search);
    if (params?.name) searchParams.set('name', params.name);
    if (params?.givenName) searchParams.set('givenName', params.givenName);
    if (params?.familyName) searchParams.set('familyName', params.familyName);
    if (params?.tradingName) searchParams.set('tradingName', params.tradingName);
    if (params?.type) searchParams.set('type', params.type);
    if (params?.status) searchParams.set('status', params.status);

    const response = await apiClient.get<PartyUnion[]>(`/api/parties?${searchParams.toString()}`);
    return response.data;
}

// Fetch single party by ID
async function fetchParty(id: string): Promise<PartyUnion> {
    const response = await apiClient.get<PartyUnion>(`/api/parties/${id}`);
    return response.data;
}

// Create party
async function createParty(payload: CreatePartyPayload): Promise<PartyUnion> {
    const response = await apiClient.post<PartyUnion>('/api/parties', payload);
    return response.data;
}

// Update party
async function updateParty(payload: UpdatePartyPayload): Promise<PartyUnion> {
    const response = await apiClient.put<PartyUnion>(`/api/parties/${payload.id}`, payload);
    return response.data;
}

// Delete party
async function deleteParty(id: string): Promise<void> {
    await apiClient.delete(`/api/parties/${id}`);
}

// Permanently delete a soft-deleted party
async function purgeParty(id: string): Promise<void> {
    await apiClient.delete(`/api/parties/${id}/purge`);
}

// React Query Hooks

export function useParties(params?: SearchPartyParams) {
    return useQuery({
        queryKey: [PARTIES_KEY, params],
        queryFn: () => fetchParties(params),
    });
}

export function useParty(id: string | undefined) {
    return useQuery({
        queryKey: [PARTIES_KEY, id],
        queryFn: () => fetchParty(id!),
        enabled: !!id,
    });
}

export function useCreateParty() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: createParty,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: [PARTIES_KEY] });
        },
    });
}

export function useUpdateParty() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: updateParty,
        onSuccess: (data) => {
            queryClient.invalidateQueries({ queryKey: [PARTIES_KEY] });
            queryClient.setQueryData([PARTIES_KEY, data.id], data);
        },
    });
}

export function useDeleteParty() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: deleteParty,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: [PARTIES_KEY] });
        },
    });
}

export function usePurgeParty() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: purgeParty,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: [PARTIES_KEY] });
        },
    });
}
