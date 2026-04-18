import { useState } from 'react';
import { Outlet } from 'react-router-dom';
import { Sidebar } from './Sidebar';
import { Menu } from 'lucide-react';
import { LoginPage } from '../../features/auth/LoginPage';
import { useAuth } from '../../auth/context';
import './Layout.css';

export function Layout() {
    const { isAuthenticated, isLoading } = useAuth();
    const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);
    const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-screen bg-slate-50">
                <div className="flex flex-col items-center gap-4">
                    <div className="w-12 h-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
                    <p className="text-sm text-slate-500 font-medium">Loading session...</p>
                </div>
            </div>
        );
    }

    if (!isAuthenticated) {
        return <LoginPage />;
    }

    const toggleSidebarCollapse = () => {
        setIsSidebarCollapsed(!isSidebarCollapsed);
    };

    const toggleMobileMenu = () => {
        setIsMobileMenuOpen(!isMobileMenuOpen);
    };

    const closeMobileMenu = () => {
        setIsMobileMenuOpen(false);
    };

    return (
        <div className="layout">
            <Sidebar
                collapsed={isSidebarCollapsed}
                mobileOpen={isMobileMenuOpen}
                onToggleCollapse={toggleSidebarCollapse}
            />
            <div className={`layout-main ${isSidebarCollapsed ? 'layout-main--collapsed' : ''}`}>
                <header className="layout-header">
                    <div className="layout-header-content">
                        <div className="layout-header-title">
                            <h1>TMF Demo Dashboard</h1>
                            <p className="layout-subtitle">Managed via Golang BFF & RabbitMQ</p>
                        </div>
                        {!isMobileMenuOpen && (
                            <button
                                className="mobile-menu-toggle p-2 hover:bg-slate-800 rounded-lg text-slate-400 hover:text-white transition-colors"
                                onClick={toggleMobileMenu}
                                aria-label="Toggle menu"
                            >
                                <Menu size={24} />
                            </button>
                        )}
                    </div>
                </header>
                <main className="layout-content">
                    <Outlet />
                </main>
                {isMobileMenuOpen && (
                    <div
                        className="mobile-overlay"
                        onClick={closeMobileMenu}
                        aria-hidden="true"
                    />
                )}
            </div>
        </div>
    );
}
