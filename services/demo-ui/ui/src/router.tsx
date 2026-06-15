import { createBrowserRouter } from 'react-router-dom';
import { Layout } from './components/layout/Layout';
import { RequireAuth } from './components/auth/RequireAuth';
import { ProvisioningGate } from './components/auth/ProvisioningGate';
import { IndexRedirect } from './components/auth/IndexRedirect';
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
const CatalogDetailPage = lazy(() => import('./features/catalog/CatalogDetailPage'));
const CatalogEditPage = lazy(() => import('./features/catalog/CatalogEditPage'));

const CategoryListPage = lazy(() => import('./features/catalog/CategoryListPage'));
const CategoryDetailPage = lazy(() => import('./features/catalog/CategoryDetailPage'));
const CategoryEditPage = lazy(() => import('./features/catalog/CategoryEditPage'));

const OfferingListPage = lazy(() => import('./features/catalog/OfferingListPage'));
const OfferingDetailPage = lazy(() => import('./features/catalog/OfferingDetailPage'));
const OfferingEditPage = lazy(() => import('./features/catalog/OfferingEditPage'));

const QualifyPage = lazy(() => import('./features/ordering/QualifyPage'));
const CartPage = lazy(() => import('./features/ordering/CartPage'));
const CheckoutPage = lazy(() => import('./features/ordering/CheckoutPage'));
const OrderStatusPage = lazy(() => import('./features/ordering/OrderStatusPage'));
const OrderConfirmationPage = lazy(() => import('./features/ordering/OrderConfirmationPage'));

const DebugConsolePage = lazy(() => import('./features/debug/DebugConsolePage'));

const CompleteProfilePage = lazy(() => import('./features/provisioning/CompleteProfilePage'));

// Loading fallback component
import { PageLoader } from './design-system/components/common/PageLoader';

export const router = createBrowserRouter([
    {
        path: '/',
        element: <Layout />,
        children: [
            // Landing redirect: authenticated -> dashboard, anonymous -> Check Availability
            {
                index: true,
                element: <IndexRedirect />,
            },
            // --- Public (anonymous-accessible): Check Availability only ---
            {
                path: 'order/qualify',
                element: (
                    <Suspense fallback={<PageLoader />}>
                        <QualifyPage />
                    </Suspense>
                ),
            },
            // --- Protected: everything else requires authentication ---
            {
                element: <RequireAuth />,
                children: [
                    // Provisioning form: authenticated but NOT behind the
                    // provisioning gate (otherwise it would redirect to itself).
                    {
                        path: 'complete-profile',
                        element: (
                            <Suspense fallback={<PageLoader />}>
                                <CompleteProfilePage />
                            </Suspense>
                        ),
                    },
                    // Everything else also requires a provisioned customer.
                    {
                        element: <ProvisioningGate />,
                        children: [
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
                                                <CatalogDetailPage />
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
                                                <CategoryDetailPage />
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
                                                <OfferingDetailPage />
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
                    // Protected ordering routes (cart, checkout, status, confirmation)
                    {
                        path: 'order',
                        children: [
                            {
                                path: 'cart',
                                element: (
                                    <Suspense fallback={<PageLoader />}>
                                        <CartPage />
                                    </Suspense>
                                ),
                            },
                            {
                                path: 'checkout',
                                element: (
                                    <Suspense fallback={<PageLoader />}>
                                        <CheckoutPage />
                                    </Suspense>
                                ),
                            },
                            {
                                path: 'status/:sagaId',
                                element: (
                                    <Suspense fallback={<PageLoader />}>
                                        <OrderStatusPage />
                                    </Suspense>
                                ),
                            },
                            {
                                path: 'confirmation/:orderId',
                                element: (
                                    <Suspense fallback={<PageLoader />}>
                                        <OrderConfirmationPage />
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
                ],
            },
        ],
    },
]);
