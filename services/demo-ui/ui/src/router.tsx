import { createBrowserRouter, Navigate } from 'react-router-dom';
import { Layout } from './components/layout/Layout';
import { lazy, Suspense } from 'react';

// Lazy load pages for code splitting
const PartyListPage = lazy(() => import('./features/parties/PartyListPage'));
const PartyDetailPage = lazy(() => import('./features/parties/PartyDetailPage'));
const PartyFormPage = lazy(() => import('./features/parties/PartyFormPage'));

const CustomerListPage = lazy(() => import('./features/customers/CustomerListPage'));
const CustomerDetailPage = lazy(() => import('./features/customers/CustomerDetailPage'));
const CustomerEditPage = lazy(() => import('./features/customers/CustomerEditPage'));
const CustomerOnboardPage = lazy(() => import('./features/customers/CustomerOnboardPage'));

const SpecificationListPage = lazy(() => import('./features/catalog/SpecificationListPage'));
const SpecificationDetailPage = lazy(() => import('./features/catalog/SpecificationDetailPage'));
const SpecificationEditPage = lazy(() => import('./features/catalog/SpecificationEditPage'));

const CatalogListPage = lazy(() => import('./features/catalog/CatalogListPage'));
const CatalogEditPage = lazy(() => import('./features/catalog/CatalogEditPage'));

const CategoryListPage = lazy(() => import('./features/catalog/CategoryListPage'));
const CategoryEditPage = lazy(() => import('./features/catalog/CategoryEditPage'));

const OfferingListPage = lazy(() => import('./features/catalog/OfferingListPage'));
const OfferingEditPage = lazy(() => import('./features/catalog/OfferingEditPage'));

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
            // Catalog routes
            {
                path: 'catalog',
                children: [
                    {
                        path: 'catalogs',
                        children: [
                            {
                                index: true,
                                element: (
                                    <Suspense fallback={<PageLoader />}>
                                        <CatalogListPage />
                                    </Suspense>
                                ),
                            },
                            {
                                path: 'new',
                                element: (
                                    <Suspense fallback={<PageLoader />}>
                                        <CatalogEditPage />
                                    </Suspense>
                                ),
                            },
                            {
                                path: ':id',
                                element: (
                                    <Suspense fallback={<PageLoader />}>
                                        <CatalogEditPage />
                                    </Suspense>
                                ),
                            },
                            {
                                path: ':id/edit',
                                element: (
                                    <Suspense fallback={<PageLoader />}>
                                        <CatalogEditPage />
                                    </Suspense>
                                ),
                            },
                        ]
                    },
                    {
                        path: 'categories',
                        children: [
                            {
                                index: true,
                                element: (
                                    <Suspense fallback={<PageLoader />}>
                                        <CategoryListPage />
                                    </Suspense>
                                ),
                            },
                            {
                                path: 'new',
                                element: (
                                    <Suspense fallback={<PageLoader />}>
                                        <CategoryEditPage />
                                    </Suspense>
                                ),
                            },
                            {
                                path: ':id',
                                element: (
                                    <Suspense fallback={<PageLoader />}>
                                        <CategoryEditPage />
                                    </Suspense>
                                ),
                            },
                            {
                                path: ':id/edit',
                                element: (
                                    <Suspense fallback={<PageLoader />}>
                                        <CategoryEditPage />
                                    </Suspense>
                                ),
                            },
                        ],
                    },
                    {
                        path: 'specifications',
                        children: [
                            {
                                index: true,
                                element: (
                                    <Suspense fallback={<PageLoader />}>
                                        <SpecificationListPage />
                                    </Suspense>
                                ),
                            },
                            {
                                path: 'new',
                                element: (
                                    <Suspense fallback={<PageLoader />}>
                                        <SpecificationEditPage />
                                    </Suspense>
                                ),
                            },
                            {
                                path: ':id',
                                element: (
                                    <Suspense fallback={<PageLoader />}>
                                        <SpecificationDetailPage />
                                    </Suspense>
                                ),
                            },
                            {
                                path: ':id/edit',
                                element: (
                                    <Suspense fallback={<PageLoader />}>
                                        <SpecificationEditPage />
                                    </Suspense>
                                ),
                            },
                        ],
                    },
                    {
                        path: 'offerings',
                        children: [
                            {
                                index: true,
                                element: (
                                    <Suspense fallback={<PageLoader />}>
                                        <OfferingListPage />
                                    </Suspense>
                                ),
                            },
                            {
                                path: 'new',
                                element: (
                                    <Suspense fallback={<PageLoader />}>
                                        <OfferingEditPage />
                                    </Suspense>
                                ),
                            },
                            {
                                path: ':id',
                                element: (
                                    <Suspense fallback={<PageLoader />}>
                                        <OfferingEditPage />
                                    </Suspense>
                                ),
                            },
                            {
                                path: ':id/edit',
                                element: (
                                    <Suspense fallback={<PageLoader />}>
                                        <OfferingEditPage />
                                    </Suspense>
                                ),
                            },
                        ],
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
