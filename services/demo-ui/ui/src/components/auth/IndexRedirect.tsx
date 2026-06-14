import { Navigate } from 'react-router-dom';
import { useAuth } from '../../auth/context';
import { PageLoader } from '../../design-system/components/common/PageLoader';

/**
 * Landing redirect for the root path. Authenticated users go to the management
 * dashboard; anonymous users go to the public Check Availability page.
 */
export function IndexRedirect() {
    const { isAuthenticated, isLoading } = useAuth();

    if (isLoading) {
        return <PageLoader />;
    }

    return <Navigate to={isAuthenticated ? '/parties' : '/order/qualify'} replace />;
}
