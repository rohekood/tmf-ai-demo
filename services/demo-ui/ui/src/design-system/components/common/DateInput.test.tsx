import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { createRef } from 'react';
import { fireEvent } from '@testing-library/react';
import { DateInput } from './DateInput';

describe('DateInput', () => {
    it('renders a native date input with the date-input class', () => {
        render(<DateInput aria-label="date" />);
        const input = screen.getByLabelText('date');
        expect(input).toHaveAttribute('type', 'date');
        expect(input).toHaveClass('date-input');
    });

    it('renders a label associated with the input when label is provided', () => {
        render(<DateInput label="Start Date" />);
        const input = screen.getByLabelText('Start Date');
        expect(input).toBeInTheDocument();
        // label htmlFor matches the input id
        expect(input.id).toBeTruthy();
    });

    it('does not render a label wrapper when no label is given', () => {
        const { container } = render(<DateInput aria-label="bare" />);
        expect(container.querySelector('label')).toBeNull();
        expect(container.querySelector('.form-group')).toBeNull();
    });

    it('forwards value and change events', () => {
        const onChange = vi.fn();
        render(<DateInput label="Pick" value="2026-06-14" onChange={onChange} />);
        const input = screen.getByLabelText('Pick') as HTMLInputElement;
        expect(input.value).toBe('2026-06-14');
        fireEvent.change(input, { target: { value: '2026-07-01' } });
        expect(onChange).toHaveBeenCalled();
    });

    it('respects a provided id and forwards a ref', () => {
        const ref = createRef<HTMLInputElement>();
        render(<DateInput id="my-date" label="Custom" ref={ref} />);
        const input = screen.getByLabelText('Custom');
        expect(input.id).toBe('my-date');
        expect(ref.current).toBe(input);
    });

    it('supports disabled state', () => {
        render(<DateInput label="Off" disabled />);
        expect(screen.getByLabelText('Off')).toBeDisabled();
    });
});
