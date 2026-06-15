import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { useProvisioning } from '../../features/provisioning/useProvisioning';
import { PageLoader } from '../../design-system/components/common/PageLoader';

/**
 * Ensures the authenticated user is backed by a Customer before reaching the
 * protected app. If provisioning is still required, the user is redirected to
 * the "Complete your profile" form (carrying the intended destination so they
 * can resume afterwards). Assumes it is nested inside <RequireAuth>.
 */
export function ProvisioningGate() {
    const { isLoading, isError, status } = useProvisioning();
    const location = useLocation();

    if (isLoading) {
        return <PageLoader />;
    }

    // status === 'ready' means a customer already exists. Anything else — needs
    // provisioning, or the resolve failed (e.g. no email claim available) — sends
    // the user to the profile form, which can complete provisioning directly.
    if (status !== 'ready' || isError) {
        return (
            <Navigate
                to="/complete-profile"
                replace
                state={{ returnTo: location.pathname + location.search }}
            />
        );
    }

    return <Outlet />;
}
