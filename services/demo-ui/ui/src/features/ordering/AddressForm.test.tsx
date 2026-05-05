import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { AddressForm } from './AddressForm';
import type { AddressFormData } from './AddressForm';

const defaultAddress: AddressFormData = {
    street: '',
    number: '',
    city: '',
    zip: '',
};

describe('AddressForm', () => {
    it('renders all four required fields with labels', () => {
        render(
            <AddressForm
                address={defaultAddress}
                onChange={vi.fn()}
                onSubmit={vi.fn()}
                isPending={false}
            />
        );

        expect(screen.getByLabelText(/Street/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/Number/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/City/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/ZIP/i)).toBeInTheDocument();
    });

    it('all fields have required attribute', () => {
        render(
            <AddressForm
                address={defaultAddress}
                onChange={vi.fn()}
                onSubmit={vi.fn()}
                isPending={false}
            />
        );

        expect(screen.getByLabelText(/Street/i)).toBeRequired();
        expect(screen.getByLabelText(/Number/i)).toBeRequired();
        expect(screen.getByLabelText(/City/i)).toBeRequired();
        expect(screen.getByLabelText(/ZIP/i)).toBeRequired();
    });

    it('shows "Check Availability" button when not pending', () => {
        render(
            <AddressForm
                address={defaultAddress}
                onChange={vi.fn()}
                onSubmit={vi.fn()}
                isPending={false}
            />
        );

        const button = screen.getByRole('button');
        expect(button).toHaveTextContent('Check Availability');
        expect(button).not.toBeDisabled();
    });

    it('shows "Checking..." and disables button when isPending is true', () => {
        render(
            <AddressForm
                address={defaultAddress}
                onChange={vi.fn()}
                onSubmit={vi.fn()}
                isPending={true}
            />
        );

        const button = screen.getByRole('button');
        expect(button).toHaveTextContent('Checking...');
        expect(button).toBeDisabled();
    });

    it('calls onChange with updated street value', async () => {
        const user = userEvent.setup();
        const handleChange = vi.fn();

        render(
            <AddressForm
                address={defaultAddress}
                onChange={handleChange}
                onSubmit={vi.fn()}
                isPending={false}
            />
        );

        await user.type(screen.getByLabelText(/Street/i), 'Main St');
        expect(handleChange).toHaveBeenCalledWith(expect.objectContaining({ street: 'M' }));
    });

    it('calls onChange with updated number value', async () => {
        const user = userEvent.setup();
        const handleChange = vi.fn();

        render(
            <AddressForm
                address={defaultAddress}
                onChange={handleChange}
                onSubmit={vi.fn()}
                isPending={false}
            />
        );

        await user.type(screen.getByLabelText(/Number/i), '42');
        expect(handleChange).toHaveBeenCalledWith(expect.objectContaining({ number: '4' }));
    });

    it('calls onChange with updated city value', async () => {
        const user = userEvent.setup();
        const handleChange = vi.fn();

        render(
            <AddressForm
                address={defaultAddress}
                onChange={handleChange}
                onSubmit={vi.fn()}
                isPending={false}
            />
        );

        await user.type(screen.getByLabelText(/City/i), 'Tallinn');
        expect(handleChange).toHaveBeenCalledWith(expect.objectContaining({ city: 'T' }));
    });

    it('calls onChange with updated zip value', async () => {
        const user = userEvent.setup();
        const handleChange = vi.fn();

        render(
            <AddressForm
                address={defaultAddress}
                onChange={handleChange}
                onSubmit={vi.fn()}
                isPending={false}
            />
        );

        await user.type(screen.getByLabelText(/ZIP/i), '10001');
        expect(handleChange).toHaveBeenCalledWith(expect.objectContaining({ zip: '1' }));
    });

    it('calls onSubmit when form is submitted', async () => {
        const user = userEvent.setup();
        const handleSubmit = vi.fn((e) => e.preventDefault());

        render(
            <AddressForm
                address={{ street: 'Main St', number: '1', city: 'Tallinn', zip: '10001' }}
                onChange={vi.fn()}
                onSubmit={handleSubmit}
                isPending={false}
            />
        );

        await user.click(screen.getByRole('button'));
        expect(handleSubmit).toHaveBeenCalled();
    });
});
