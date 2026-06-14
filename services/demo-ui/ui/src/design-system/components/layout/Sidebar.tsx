import React, { type ReactNode } from 'react';
import { ChevronRight, ChevronsLeft, ChevronsRight, LogOut, LogIn } from 'lucide-react';
import './Sidebar.css';

export interface NavItemDef {
    path: string;
    label: string;
    icon: ReactNode;
    title?: string;
    isActive?: boolean;
}

export interface UserProfile {
    name?: string;
    email?: string;
    picture?: string;
}

export interface SidebarProps {
    /**
     * Branding/App Name.
     */
    appName?: string;

    /**
     * Whether the sidebar is collapsed.
     */
    collapsed: boolean;

    /**
     * Whether the mobile menu is open.
     */
    mobileOpen?: boolean;

    /**
     * Callback to toggle collapse state.
     */
    onToggleCollapse: () => void;

    /**
     * Groups of navigation items. keys can be section titles.
     */
    navGroups: Array<{
        title?: string;
        items: NavItemDef[];
    }>;

    /**
     * Function to render a navigation link (allows using router Link/NavLink).
     */
    renderLink: (item: NavItemDef, children: ReactNode, className: string) => ReactNode;

    /**
     * User profile information.
     */
    user?: UserProfile;

    /**
     * Is user authenticated?
     */
    isAuthenticated?: boolean;

    /**
     * Is auth data loading?
     */
    isLoading?: boolean;

    /**
     * Login/Logout callbacks.
     */
    onLogin?: () => void;
    onLogout?: () => void;

    /**
     * Class name.
     */
    className?: string;
}

export const Sidebar: React.FC<SidebarProps> = ({
    appName = 'Demo UI',
    collapsed,
    mobileOpen = false,
    onToggleCollapse,
    navGroups,
    renderLink,
    user,
    isAuthenticated = false,
    isLoading = false,
    onLogin,
    onLogout,
    className = '',
}) => {
    return (
        <aside className={`sidebar ${collapsed ? 'sidebar--collapsed' : ''} ${mobileOpen ? 'sidebar--open' : ''} ${className}`}>
            {/* Brand */}
            <div className="sidebar-brand">
                <div className="sidebar-logo">
                    <span className="sidebar-logo-text">TMF</span>
                </div>
                {!collapsed && <span className="sidebar-brand-name">{appName}</span>}
                {collapsed && <span className="sidebar-brand-name sr-only">{appName}</span>}
            </div>

            {/* Nav */}
            <nav className="sidebar-nav">
                {navGroups.map((group, idx) => (
                    <div key={idx} className="sidebar-nav-section">
                        {!collapsed && group.title && (
                            <h3 className="sidebar-nav-title">
                                {group.title}
                            </h3>
                        )}
                        <div className="sidebar-nav-group">
                            {group.items.map((item) => {
                                const linkClass = `sidebar-nav-item ${item.isActive ? 'sidebar-nav-item--active' : ''}`;

                                const content = (
                                    <>
                                        <span className="sidebar-nav-icon">
                                            {item.icon}
                                        </span>
                                        {!collapsed && <span className="sidebar-nav-label">{item.label}</span>}
                                        {!collapsed && <ChevronRight size={16} className="sidebar-nav-arrow" />}
                                    </>
                                );

                                return (
                                    <div key={item.path} title={collapsed ? item.label : undefined}>
                                        {renderLink(item, content, linkClass)}
                                    </div>
                                );
                            })}
                        </div>
                    </div>
                ))}
            </nav>

            {/* Footer */}
            <div className="sidebar-footer">
                <div className="sidebar-user-section">
                    {!isLoading && (
                        isAuthenticated ? (
                            <div className="sidebar-user-card">
                                <div className={`sidebar-user-avatar-wrapper ${collapsed ? 'sidebar-user-avatar-wrapper--center' : ''}`}>
                                    <div className="sidebar-user-avatar">
                                        {user?.picture ? (
                                            <img src={user.picture} alt={user.name} className="sidebar-user-avatar-img" />
                                        ) : (
                                            <span className="sidebar-user-avatar-fallback">
                                                {user?.name?.charAt(0) || 'U'}
                                            </span>
                                        )}
                                    </div>
                                </div>

                                {!collapsed && (
                                    <>
                                        <div className="sidebar-user-info">
                                            <div className="sidebar-user-name">{user?.name}</div>
                                            <div className="sidebar-user-email">{user?.email}</div>
                                        </div>
                                        {onLogout && (
                                            <button
                                                onClick={onLogout}
                                                className="sidebar-logout-btn"
                                                title="Log out"
                                            >
                                                <LogOut size={16} />
                                            </button>
                                        )}
                                    </>
                                )}
                            </div>
                        ) : (
                            !collapsed && onLogin && (
                                <div className="sidebar-login-prompt">
                                    <p className="sidebar-login-text">Please login to continue</p>
                                    <button
                                        onClick={onLogin}
                                        className="btn-primary"
                                        style={{ width: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '8px', border: 'none', cursor: 'pointer', padding: '0.5rem' }}
                                    >
                                        <LogIn size={16} />
                                        <span>Log In</span>
                                    </button>
                                </div>
                            )
                        )
                    )}
                </div>

                <div className={`sidebar-status ${collapsed ? 'sidebar-status--center' : ''}`}>
                    {!collapsed && (
                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                            <span className="sidebar-status-dot sidebar-status-dot--connected"></span>
                            <span>Connected</span>
                        </div>
                    )}

                    <button
                        className="sidebar-collapse-toggle"
                        onClick={onToggleCollapse}
                        title={collapsed ? "Expand Sidebar" : "Collapse Sidebar"}
                    >
                        {collapsed ? <ChevronsRight size={18} /> : <ChevronsLeft size={18} />}
                    </button>
                </div>
            </div>
        </aside>
    );
};
