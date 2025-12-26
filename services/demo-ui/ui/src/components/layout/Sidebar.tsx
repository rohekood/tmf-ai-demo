import { NavLink } from 'react-router-dom';
import { Users, Building2, Bug, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-react';
import { useAuth0 } from "@auth0/auth0-react";
import { LoginButton } from '../auth/LoginButton';
import { LogoutButton } from '../auth/LogoutButton';
import './Sidebar.css';

interface NavItemProps {
    to: string;
    icon: React.ReactNode;
    label: string;
    title?: string;
}

function NavItem({ to, icon, label, title }: NavItemProps) {
    return (
        <NavLink
            to={to}
            title={title}
            className={({ isActive }) =>
                `sidebar-nav-item ${isActive ? 'sidebar-nav-item--active' : ''} ${!label ? 'justify-center px-0' : ''}`
            }
        >
            <span className="sidebar-nav-icon">{icon}</span>
            {label && <span className="sidebar-nav-label">{label}</span>}
            {label && <ChevronRight className="sidebar-nav-arrow" size={16} />}
        </NavLink>
    );
}

interface SidebarProps {
    collapsed: boolean;
    mobileOpen: boolean;
    onToggleCollapse: () => void;
}

export function Sidebar({ collapsed, mobileOpen, onToggleCollapse }: SidebarProps) {
    const { user, isAuthenticated, isLoading } = useAuth0();

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
                                <div className={`sidebar-user-avatar-wrapper ${collapsed ? 'mx-auto' : ''}`}>
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

                <div className={`sidebar-status ${collapsed ? 'justify-center' : ''}`} role="status">
                    <span className="sidebar-status-dot sidebar-status-dot--connected"></span>
                    {!collapsed && <span className="sidebar-status-text">Connected</span>}
                </div>

                <button
                    className="sidebar-collapse-toggle hidden md:flex"
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
