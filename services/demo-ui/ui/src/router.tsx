import { createBrowserRouter, Navigate } from 'react-router-dom';
import { Layout } from './components/layout/Layout';
import { lazy, Suspense } from 'react';

// Lazy load pages for code splitting
const PartyListPage = lazy(() => import('./features/parties/PartyListPage'));
const PartyDetailPage = lazy(() => import('./features/parties/PartyDetailPage'));
const PartyFormPage = lazy(() => import('./features/parties/PartyFormPage'));

const CustomerListPage = lazy(() => import('./features/customers/CustomerListPage'));
const CustomerDetailPage = lazy(() => import('./features/customers/CustomerDetailPage'));
const CustomerOnboardPage = lazy(() => import('./features/customers/CustomerOnboardPage'));
const CustomerEditPage = lazy(() => import('./features/customers/CustomerEditPage'));

const DebugConsolePage = lazy(() => import('./features/debug/DebugConsolePage'));

// Loading fallback component
import { PageLoader } from './design-system/components/common/PageLoader';

export const router = createBrowserRouter([
    {
        path: '/',
        element: <Layout />,
        children: [
            // Default redirect to parties
            {
                index: true,
                element: <Navigate to="/parties" replace />,
            },
            // Party routes
            {
                path: 'parties',
                children: [
                    {
                        index: true,
                        element: (
                            <Suspense fallback={<PageLoader />}>
                                <PartyListPage />
                            </Suspense>
                        ),
                    },
                    {
                        path: 'new',
                        element: (
                            <Suspense fallback={<PageLoader />}>
                                <PartyFormPage />
                            </Suspense>
                        ),
                    },
                    {
                        path: ':id',
                        element: (
                            <Suspense fallback={<PageLoader />}>
                                <PartyDetailPage />
                            </Suspense>
                        ),
                    },
                    {
                        path: ':id/edit',
                        element: (
                            <Suspense fallback={<PageLoader />}>
                                <PartyFormPage />
                            </Suspense>
                        ),
                    },
                ],
            },
            // Customer routes
            {
                path: 'customers',
                children: [
                    {
                        index: true,
                        element: (
                            <Suspense fallback={<PageLoader />}>
                                <CustomerListPage />
                            </Suspense>
                        ),
                    },
                    {
                        path: 'new',
                        element: (
                            <Suspense fallback={<PageLoader />}>
                                <CustomerOnboardPage />
                            </Suspense>
                        ),
                    },
                    {
                        path: ':id',
                        element: (
                            <Suspense fallback={<PageLoader />}>
                                <CustomerDetailPage />
                            </Suspense>
                        ),
                    },
                    {
                        path: ':id/edit',
                        element: (
                            <Suspense fallback={<PageLoader />}>
                                <CustomerEditPage />
                            </Suspense>
                        ),
                    },
                ],
            },
            // Debug route
            {
                path: 'debug',
                element: (
                    <Suspense fallback={<PageLoader />}>
                        <DebugConsolePage />
                    </Suspense>
                ),
            },
        ],
    },
]);
