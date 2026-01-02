import React, { type ReactNode } from 'react';
import './IconButtonArea.css';

export interface IconButtonAreaProps {
    /**
     * The alignment of the buttons within the container.
     * @default 'end'
     */
    alignment?: 'start' | 'center' | 'end' | 'between';

    /**
     * The buttons to be displayed.
     */
    children: ReactNode;

    /**
     * Optional additional CSS classes.
     */
    className?: string;
}

/**
 * IconButtonArea is a container for icon-only action buttons ensuring consistent spacing and alignment.
 */
export const IconButtonArea: React.FC<IconButtonAreaProps> = ({
    alignment = 'end',
    children,
    className = '',
}) => {
    return (
        <div className={`icon-button-area icon-button-area--${alignment} ${className}`}>
            {children}
        </div>
    );
};
