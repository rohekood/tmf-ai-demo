import { NavLink } from 'react-router-dom';
import { Users, Building2, Bug, ChevronRight } from 'lucide-react';
import './Sidebar.css';

interface NavItemProps {
    to: string;
    icon: React.ReactNode;
    label: string;
}

function NavItem({ to, icon, label }: NavItemProps) {
    return (
        <NavLink
            to={to}
            className={({ isActive }) =>
                `sidebar-nav-item ${isActive ? 'sidebar-nav-item--active' : ''}`
            }
        >
            <span className="sidebar-nav-icon">{icon}</span>
            <span className="sidebar-nav-label">{label}</span>
            <ChevronRight className="sidebar-nav-arrow" size={16} />
        </NavLink>
    );
}

export function Sidebar() {
    return (
        <aside className="sidebar">
            <div className="sidebar-brand">
                <div className="sidebar-logo">
                    <span className="sidebar-logo-text">TMF</span>
                </div>
                <span className="sidebar-brand-name">Demo UI</span>
            </div>

            <nav className="sidebar-nav">
                <div className="sidebar-nav-section">
                    <h3 className="sidebar-nav-title">Management</h3>
                    <NavItem
                        to="/parties"
                        icon={<Users size={20} />}
                        label="Parties"
                    />
                    <NavItem
                        to="/customers"
                        icon={<Building2 size={20} />}
                        label="Customers"
                    />
                </div>

                <div className="sidebar-nav-section">
                    <h3 className="sidebar-nav-title">Developer</h3>
                    <NavItem
                        to="/debug"
                        icon={<Bug size={20} />}
                        label="Debug Console"
                    />
                </div>
            </nav>

            <div className="sidebar-footer">
                <div className="sidebar-status" role="status">
                    <span className="sidebar-status-dot sidebar-status-dot--connected"></span>
                    <span className="sidebar-status-text">Connected</span>
                </div>
            </div>
        </aside>
    );
}
