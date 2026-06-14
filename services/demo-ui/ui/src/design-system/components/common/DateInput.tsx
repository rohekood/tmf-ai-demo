import React, { forwardRef, useId } from 'react';
import './DateInput.css';

export interface DateInputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'type'> {
    /**
     * Optional field label. When provided, the input is wrapped in a
     * `.form-group` together with a `<label>`. When omitted, only the input
     * element is rendered (so it can be placed inside an existing field group).
     */
    label?: string;
}

/**
 * A reusable date picker input styled for the dark theme.
 *
 * Native `<input type="date">` is not covered by the global input styles and
 * its calendar indicator renders black (invisible in dark mode). This component
 * applies the dark field styling, recolours the calendar icon to white, and
 * shows a pointer cursor when hovering the icon. Values use the native ISO
 * (`yyyy-MM-dd`) format; display-only dates are formatted via `lib/date`.
 */
export const DateInput = forwardRef<HTMLInputElement, DateInputProps>(
    ({ label, id, className = '', ...props }, ref) => {
        const generatedId = useId();
        const inputId = id ?? generatedId;

        const input = (
            <input
                id={inputId}
                type="date"
                lang="et-EE"
                ref={ref}
                className={`date-input ${className}`.trim()}
                {...props}
            />
        );

        if (!label) return input;

        return (
            <div className="form-group">
                <label htmlFor={inputId}>{label}</label>
                {input}
            </div>
        );
    }
);

DateInput.displayName = 'DateInput';
