import React from 'react';
import './Footer.css';

export interface FooterProps {
    /**
     * The application version.
     */
    version: string;

    /**
     * The company or copyright holder name.
     */
    companyName?: string;

    /**
     * Links to validation/support/docs.
     */
    links?: Array<{ label: string; href: string }>;

    /**
     * Optional custom class name.
     */
    className?: string;
}

/**
 * Standard application footer.
 */
export const Footer: React.FC<FooterProps> = ({
    version,
    companyName = 'TMF Demo',
    links = [],
    className = '',
}) => {
    const currentYear = new Date().getFullYear();

    return (
        <footer className={`footer ${className}`}>
            <div className="footer-info">
                <span>&copy; {currentYear} {companyName}</span>
                <span className="footer-divider">|</span>
                <span>Version {version}</span>
            </div>

            {links.length > 0 && (
                <div className="footer-links">
                    {links.map((link) => (
                        <a
                            key={link.label}
                            href={link.href}
                            className="footer-link"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            {link.label}
                        </a>
                    ))}
                </div>
            )}
        </footer>
    );
};
