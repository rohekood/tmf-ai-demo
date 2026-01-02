import React, { type ReactNode } from 'react';
import { Menu } from 'lucide-react';
import './TopBar.css';

export interface TopBarProps {
    /**
     * The main title of the application or page.
     */
    title: string;

    /**
     * Subtitle or description text.
     */
    subtitle?: string;

    /**
     * Whether the mobile menu is open.
     */
    isMobileMenuOpen?: boolean;

    /**
     * Callback to toggle the mobile menu.
     */
    onToggleMobileMenu?: () => void;

    /**
     * Optional right-side actions (e.g., user profile, notifications).
     */
    actions?: ReactNode;

    /**
     * Optional custom class name.
     */
    className?: string;
}

/**
 * TopBar component for application header.
 */
export const TopBar: React.FC<TopBarProps> = ({
    title,
    subtitle,
    isMobileMenuOpen = false,
    onToggleMobileMenu,
    actions,
    className = '',
}) => {
    return (
        <header className={`topbar ${className}`}>
            <div className="topbar-content">
                <div className="topbar-title-section">
                    <h1 className="topbar-title">{title}</h1>
                    {subtitle && <p className="topbar-subtitle">{subtitle}</p>}
                </div>

                {actions && <div className="topbar-actions">{actions}</div>}

                {onToggleMobileMenu && !isMobileMenuOpen && (
                    <button
                        className="topbar-mobile-toggle"
                        onClick={onToggleMobileMenu}
                        aria-label="Toggle menu"
                    >
                        <Menu size={24} />
                    </button>
                )}
            </div>
        </header>
    );
};
