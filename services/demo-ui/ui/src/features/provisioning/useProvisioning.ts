import { useAuth } from '../../auth/context';
import { useResolveCustomer } from './api';

/**
 * Resolves the provisioning state of the logged-in user. For anonymous users it
 * stays idle (the query is disabled) and reports `status` as undefined.
 */
export function useProvisioning() {
    const { isAuthenticated } = useAuth();
    const query = useResolveCustomer(isAuthenticated);

    return {
        isAuthenticated,
        isLoading: isAuthenticated && query.isLoading,
        isError: query.isError,
        status: query.data?.status,
        customerId: query.data?.customer?.id,
        refetch: query.refetch,
    };
}
