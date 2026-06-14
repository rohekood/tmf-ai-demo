import { useEffect } from 'react';
import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { useAuth } from '../../auth/context';
import { PageLoader } from '../../design-system/components/common/PageLoader';

/**
 * Route guard for authenticated-only sections. Anonymous visitors are sent into
 * the login flow (preserving the page they tried to reach). When Auth0 is not
 * available (e.g. unconfigured/insecure origin) there is no way to authenticate,
 * so we fall back to the public Check Availability page instead of looping.
 */
export function RequireAuth() {
    const { isAuthenticated, isLoading, enabled, loginWithRedirect } = useAuth();
    const location = useLocation();

    useEffect(() => {
        if (!isLoading && !isAuthenticated && enabled) {
            void loginWithRedirect({
                appState: { returnTo: location.pathname + location.search },
            });
        }
    }, [isLoading, isAuthenticated, enabled, loginWithRedirect, location]);

    if (isLoading) {
        return <PageLoader />;
    }

    if (isAuthenticated) {
        return <Outlet />;
    }

    // Anonymous: if we can authenticate, show a loader while the redirect kicks
    // in; otherwise bounce to the public ordering page.
    return enabled ? <PageLoader /> : <Navigate to="/order/qualify" replace />;
}
