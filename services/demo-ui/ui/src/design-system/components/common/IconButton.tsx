import React, { forwardRef } from 'react';
import { Link, type LinkProps } from 'react-router-dom';
import './IconButton.css';

export interface IconButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    /**
     * The icon to display.
     */
    icon: React.ReactNode;

    /**
     * Visual variant of the button.
     * @default 'default'
     */
    variant?: 'default' | 'ghost' | 'danger' | 'primary';

    /**
     * Size of the button.
     * @default 'md'
     */
    size?: 'sm' | 'md' | 'lg';

    /**
     * Optional tooltip text.
     */
    title?: string;

    /**
     * If provided, renders as a Link component.
     */
    to?: string;
}

/**
 * A specialized button component for displaying icons.
 * Can render as a button or a Link if 'to' prop is provided.
 */
export const IconButton = forwardRef<HTMLButtonElement | HTMLAnchorElement, IconButtonProps>(({
    icon,
    variant = 'default',
    size = 'md',
    className = '',
    to,
    ...props
}, ref) => {
    const classes = `icon-btn icon-btn--${variant} icon-btn--${size} ${className}`;

    if (to) {
        // When 'to' is present, we render a Link. 
        // We need to cast props to avoid TS errors because we are mixing button/anchor props.
        // In a strict world we might separate the types, but this is convenient.
        return (
            <Link
                to={to}
                className={classes}
                title={props.title}
                {...(props as Omit<LinkProps, 'to'>)}
                ref={ref as React.Ref<HTMLAnchorElement>}
            >
                {icon}
            </Link>
        );
    }

    return (
        <button
            type="button"
            className={classes}
            ref={ref as React.Ref<HTMLButtonElement>}
            {...props}
        >
            {icon}
        </button>
    );
});

IconButton.displayName = 'IconButton';
