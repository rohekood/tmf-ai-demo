import { NavLink } from 'react-router-dom';
import { Users, Building2, Bug, ChevronRight, ChevronsLeft, ChevronsRight, BookOpen, Package, ShoppingCart, FolderTree, ShoppingBag, MapPin } from 'lucide-react';
import { LoginButton } from '../auth/LoginButton';
import { LogoutButton } from '../auth/LogoutButton';
import { useAuth } from '../../auth/context';
import { CartBadge } from '../../features/ordering/CartBadge';
import '../../design-system/components/layout/Sidebar.css';

interface NavItemProps {
    to: string;
    icon: React.ReactNode;
    label: string;
    title?: string;
    badge?: React.ReactNode;
}

function NavItem({ to, icon, label, title, badge }: NavItemProps) {
    return (
        <NavLink
            to={to}
            title={title}
            className={({ isActive }) =>
                `sidebar-nav-item ${isActive ? 'sidebar-nav-item--active' : ''}`
            }
        >
            <span className="sidebar-nav-icon" aria-hidden="true">{icon}</span>
            {label && <span className="sidebar-nav-label">{label}</span>}
            {badge}
            {label && <ChevronRight className="sidebar-nav-arrow" size={16} aria-hidden="true" />}
        </NavLink>
    );
}

interface SidebarProps {
    collapsed: boolean;
    mobileOpen: boolean;
    onToggleCollapse: () => void;
}

export function Sidebar({ collapsed, mobileOpen, onToggleCollapse }: SidebarProps) {
    const { user, isAuthenticated, isLoading } = useAuth();

    return (
        <aside className={`sidebar ${collapsed ? 'sidebar--collapsed' : ''} ${mobileOpen ? 'sidebar--open' : ''}`}>
            <div className="sidebar-brand">
                <div className="sidebar-logo">
                    <span className="sidebar-logo-text">TMF</span>
                </div>
                {!collapsed && <span className="sidebar-brand-name">Demo UI</span>}
                {collapsed && <span className="sidebar-brand-name sr-only">Demo UI</span>}
            </div>

            <nav className="sidebar-nav">
                <div className="sidebar-nav-section">
                    {!collapsed && <h3 className="sidebar-nav-title">Management</h3>}
                    <NavItem
                        to="/parties"
                        icon={<Users size={20} />}
                        label={collapsed ? "" : "Parties"}
                        title={collapsed ? "Parties" : undefined}
                    />
                    <NavItem
                        to="/customers"
                        icon={<Building2 size={20} />}
                        label={collapsed ? "" : "Customers"}
                        title={collapsed ? "Customers" : undefined}
                    />
                </div>

                <div className="sidebar-nav-section">
                    {!collapsed && <h3 className="sidebar-nav-title">Product Catalog</h3>}
                    <NavItem
                        to="/catalog/catalogs"
                        icon={<BookOpen size={20} />}
                        label={collapsed ? "" : "Catalogs"}
                        title={collapsed ? "Catalogs" : undefined}
                    />
                    <NavItem
                        to="/catalog/categories"
                        icon={<FolderTree size={20} />}
                        label={collapsed ? "" : "Categories"}
                        title={collapsed ? "Categories" : undefined}
                    />
                    <NavItem
                        to="/catalog/specifications"
                        icon={<Package size={20} />}
                        label={collapsed ? "" : "Specifications"}
                        title={collapsed ? "Specifications" : undefined}
                    />
                    <NavItem
                        to="/catalog/offerings"
                        icon={<ShoppingCart size={20} />}
                        label={collapsed ? "" : "Offerings"}
                        title={collapsed ? "Offerings" : undefined}
                    />
                </div>

                <div className="sidebar-nav-section">
                    {!collapsed && <h3 className="sidebar-nav-title">Ordering</h3>}
                    <NavItem
                        to="/order/qualify"
                        icon={<MapPin size={20} />}
                        label={collapsed ? "" : "Check Availability"}
                        title={collapsed ? "Check Availability" : undefined}
                    />
                    <NavItem
                        to="/order/cart"
                        icon={<ShoppingBag size={20} />}
                        label={collapsed ? "" : "Shopping Cart"}
                        title={collapsed ? "Shopping Cart" : undefined}
                        badge={!collapsed && <CartBadge />}
                    />
                </div>

                <div className="sidebar-nav-section">
                    {!collapsed && <h3 className="sidebar-nav-title">Developer</h3>}
                    <NavItem
                        to="/debug"
                        icon={<Bug size={20} />}
                        label={collapsed ? "" : "Debug Console"}
                        title={collapsed ? "Debug Console" : undefined}
                    />
                </div>
            </nav>

            <div className="sidebar-footer">
                <div className="sidebar-user-section">
                    {!isLoading && (
                        isAuthenticated ? (
                            <div className="sidebar-user-card">
                                <div className="sidebar-user-avatar-wrapper">
                                    <div className="sidebar-user-avatar">
                                        {user?.picture ? (
                                            <img src={user.picture} alt={user.name || 'User'} className="sidebar-user-avatar-img" />
                                        ) : (
                                            <span className="sidebar-user-avatar-fallback">
                                                {user?.name?.charAt(0) || 'U'}
                                            </span>
                                        )}
                                    </div>
                                </div>
                                {!collapsed && (
                                    <>
                                        <div className="sidebar-user-info" role="group" aria-label="User details">
                                            <span className="sidebar-user-name" title={user?.name}>{user?.name}</span>
                                            <span className="sidebar-user-email" title={user?.email}>{user?.email}</span>
                                        </div>
                                        <LogoutButton />
                                    </>
                                )}
                            </div>
                        ) : (
                            !collapsed && (
                                <div className="sidebar-login-prompt">
                                    <p className="sidebar-login-text">Please login to continue</p>
                                    <LoginButton />
                                </div>
                            )
                        )
                    )}
                </div>

                <div
                    className={`sidebar-status ${collapsed ? 'sidebar-status--center' : ''}`}
                    role="status"
                    title="API connection online"
                >
                    <span className="sidebar-status-dot sidebar-status-dot--connected"></span>
                    {!collapsed && <span className="sidebar-status-text">API online</span>}
                </div>

                <button
                    className="sidebar-collapse-toggle"
                    onClick={onToggleCollapse}
                    title={collapsed ? "Expand Sidebar" : "Collapse Sidebar"}
                >
                    {collapsed ? (
                        <ChevronsRight size={20} />
                    ) : (
                        <ChevronsLeft size={20} />
                    )}
                </button>
            </div>


        </aside>
    );
}
