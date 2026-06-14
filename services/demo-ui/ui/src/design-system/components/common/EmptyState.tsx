import type { ReactNode } from 'react';
import './EmptyState.css';

interface EmptyStateProps {
    /** Optional decorative icon (e.g. a lucide icon). Hidden from assistive tech. */
    icon?: ReactNode;
    /** Primary message, e.g. "No parties found." */
    title: string;
    /** Optional secondary line giving more context or a next step. */
    description?: string;
    /** Optional call-to-action (e.g. a Link or button). */
    action?: ReactNode;
    /** Render without the surrounding card (for use inside an existing card/panel). */
    bare?: boolean;
}

/**
 * Consistent empty-state block used across list/collection screens.
 * Renders inside a card with a centered icon, message, and optional CTA.
 */
export function EmptyState({ icon, title, description, action, bare }: EmptyStateProps) {
    return (
        <div className={`empty-state-block${bare ? '' : ' card'}`} role="status">
            {icon && (
                <span className="empty-state-block__icon" aria-hidden="true">
                    {icon}
                </span>
            )}
            <p className="empty-state-block__title">{title}</p>
            {description && <p className="empty-state-block__description">{description}</p>}
            {action && <div className="empty-state-block__action">{action}</div>}
        </div>
    );
}

export default EmptyState;
