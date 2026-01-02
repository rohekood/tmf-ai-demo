import React, { type ReactNode } from 'react';
import './ButtonArea.css';

export interface ButtonAreaProps {
    /**
     * The alignment of the buttons within the container.
     * @default 'end'
     */
    alignment?: 'start' | 'center' | 'end' | 'between';

    /**
     * The buttons (or other content) to be displayed.
     */
    children: ReactNode;

    /**
     * Optional additional CSS classes.
     */
    className?: string;
}

/**
 * ButtonArea is a container for action buttons ensuring consistent spacing and alignment.
 */
export const ButtonArea: React.FC<ButtonAreaProps> = ({
    alignment = 'end',
    children,
    className = '',
}) => {
    return (
        <div className={`button-area button-area--${alignment} ${className}`}>
            {children}
        </div>
    );
};
