import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../api/client';
import type {
    Catalog, CreateCatalogPayload,
    Category, CreateCategoryPayload,
    ProductSpecification, CreateProductSpecificationPayload,
    ProductOffering, CreateProductOfferingPayload
} from './types';

export const CATALOGS_KEY = 'catalogs';
export const CATEGORIES_KEY = 'categories';
export const SPECS_KEY = 'specifications';
export const OFFERINGS_KEY = 'offerings';

// --- Catalogs ---
async function fetchCatalogs(): Promise<Catalog[]> {
    const response = await apiClient.get<Catalog[]>('/api/catalogs');
    return response.data;
}

async function fetchCatalog(id: string): Promise<Catalog> {
    const response = await apiClient.get<Catalog>(`/api/catalogs/${id}`);
    return response.data;
}

async function createCatalog(payload: CreateCatalogPayload): Promise<Catalog> {
    const response = await apiClient.post<Catalog>('/api/catalogs', payload);
    return response.data;
}

async function updateCatalog(id: string, payload: Partial<Catalog>): Promise<Catalog> {
    const response = await apiClient.put<Catalog>(`/api/catalogs/${id}`, payload);
    return response.data;
}

async function deleteCatalog(id: string): Promise<void> {
    await apiClient.delete(`/api/catalogs/${id}`);
}

// --- Categories ---
async function fetchCategories(): Promise<Category[]> {
    const response = await apiClient.get<Category[]>('/api/categories');
    return response.data;
}

async function fetchCategory(id: string): Promise<Category> {
    const response = await apiClient.get<Category>(`/api/categories/${id}`);
    return response.data;
}

async function createCategory(payload: CreateCategoryPayload): Promise<Category> {
    const response = await apiClient.post<Category>('/api/categories', payload);
    return response.data;
}

async function updateCategory(id: string, payload: Partial<Category>): Promise<Category> {
    const response = await apiClient.put<Category>(`/api/categories/${id}`, payload);
    return response.data;
}

async function deleteCategory(id: string): Promise<void> {
    await apiClient.delete(`/api/categories/${id}`);
}

// --- Specifications ---
async function fetchSpecifications(): Promise<ProductSpecification[]> {
    const response = await apiClient.get<ProductSpecification[]>('/api/specifications');
    return response.data;
}

async function fetchSpecification(id: string): Promise<ProductSpecification> {
    const response = await apiClient.get<ProductSpecification>(`/api/specifications/${id}`);
    return response.data;
}

async function createSpecification(payload: CreateProductSpecificationPayload): Promise<ProductSpecification> {
    const response = await apiClient.post<ProductSpecification>('/api/specifications', payload);
    return response.data;
}

async function updateSpecification(id: string, payload: Partial<ProductSpecification>): Promise<ProductSpecification> {
    const response = await apiClient.put<ProductSpecification>(`/api/specifications/${id}`, payload);
    return response.data;
}

async function deleteSpecification(id: string): Promise<void> {
    await apiClient.delete(`/api/specifications/${id}`);
}

// --- Offerings ---
async function fetchOfferings(): Promise<ProductOffering[]> {
    const response = await apiClient.get<ProductOffering[]>('/api/offerings');
    return response.data;
}

async function fetchOffering(id: string): Promise<ProductOffering> {
    const response = await apiClient.get<ProductOffering>(`/api/offerings/${id}`);
    return response.data;
}

async function createOffering(payload: CreateProductOfferingPayload): Promise<ProductOffering> {
    const response = await apiClient.post<ProductOffering>('/api/offerings', payload);
    return response.data;
}

async function updateOffering(id: string, payload: Partial<ProductOffering>): Promise<ProductOffering> {
    const response = await apiClient.put<ProductOffering>(`/api/offerings/${id}`, payload);
    return response.data;
}

async function deleteOffering(id: string): Promise<void> {
    await apiClient.delete(`/api/offerings/${id}`);
}

// --- Hooks ---

export function useCatalogs() {
    return useQuery({ queryKey: [CATALOGS_KEY], queryFn: fetchCatalogs });
}

export function useCatalog(id: string | undefined) {
    return useQuery({
        queryKey: [CATALOGS_KEY, id],
        queryFn: () => fetchCatalog(id!),
        enabled: !!id
    });
}

export function useCreateCatalog() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: createCatalog,
        onSuccess: () => queryClient.invalidateQueries({ queryKey: [CATALOGS_KEY] }),
    });
}

// ... similarly for others, but let's focus on Specs as per Step 2

export function useSpecifications() {
    return useQuery({ queryKey: [SPECS_KEY], queryFn: fetchSpecifications });
}

export function useSpecification(id: string | undefined) {
    return useQuery({
        queryKey: [SPECS_KEY, id],
        queryFn: () => fetchSpecification(id!),
        enabled: !!id
    });
}

export function useCreateSpecification() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: createSpecification,
        onSuccess: () => queryClient.invalidateQueries({ queryKey: [SPECS_KEY] }),
    });
}

export function useUpdateSpecification() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (vars: { id: string; payload: Partial<ProductSpecification> }) =>
            updateSpecification(vars.id, vars.payload),
        onSuccess: (data) => {
            queryClient.invalidateQueries({ queryKey: [SPECS_KEY] });
            queryClient.setQueryData([SPECS_KEY, data.id], data);
        },
    });
}

export function useDeleteSpecification() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: deleteSpecification,
        onSuccess: () => queryClient.invalidateQueries({ queryKey: [SPECS_KEY] }),
    });
}

// Exporting remaining hooks to avoid unused warnings and prepare for next steps

export function useCatalogUpdate() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (vars: { id: string; payload: Partial<Catalog> }) => updateCatalog(vars.id, vars.payload),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: [CATALOGS_KEY] }),
    });
}

export function useCatalogDelete() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: deleteCatalog,
        onSuccess: () => queryClient.invalidateQueries({ queryKey: [CATALOGS_KEY] }),
    });
}

export function useCategories() {
    return useQuery({ queryKey: [CATEGORIES_KEY], queryFn: fetchCategories });
}

export function useCategory(id: string | undefined) {
    return useQuery({ queryKey: [CATEGORIES_KEY, id], queryFn: () => fetchCategory(id!), enabled: !!id });
}

export function useCreateCategory() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: createCategory,
        onSuccess: () => queryClient.invalidateQueries({ queryKey: [CATEGORIES_KEY] }),
    });
}

export function useUpdateCategory() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (vars: { id: string; payload: Partial<Category> }) => updateCategory(vars.id, vars.payload),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: [CATEGORIES_KEY] }),
    });
}

export function useDeleteCategory() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: deleteCategory,
        onSuccess: () => queryClient.invalidateQueries({ queryKey: [CATEGORIES_KEY] }),
    });
}

export function useOfferings() {
    return useQuery({ queryKey: [OFFERINGS_KEY], queryFn: fetchOfferings });
}

export function useOffering(id: string | undefined) {
    return useQuery({ queryKey: [OFFERINGS_KEY, id], queryFn: () => fetchOffering(id!), enabled: !!id });
}

export function useCreateOffering() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: createOffering,
        onSuccess: () => queryClient.invalidateQueries({ queryKey: [OFFERINGS_KEY] }),
    });
}

export function useUpdateOffering() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (vars: { id: string; payload: Partial<ProductOffering> }) => updateOffering(vars.id, vars.payload),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: [OFFERINGS_KEY] }),
    });
}

export function useDeleteOffering() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: deleteOffering,
        onSuccess: () => queryClient.invalidateQueries({ queryKey: [OFFERINGS_KEY] }),
    });
}
