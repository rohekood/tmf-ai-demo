import { render, screen, fireEvent } from '@testing-library/react';
import PaymentMethodsForm from './PaymentMethodsForm';
import { vi, describe, it, expect } from 'vitest';

describe('PaymentMethodsForm', () => {
    it('renders empty state correctly', () => {
        render(<PaymentMethodsForm items={[]} onChange={vi.fn()} />);
        expect(screen.getByText('No payment methods added')).toBeInTheDocument();
    });

    it('adds a new payment method', () => {
        const onChange = vi.fn();
        render(<PaymentMethodsForm items={[]} onChange={onChange} />);

        fireEvent.click(screen.getByText('Add Method'));

        expect(onChange).toHaveBeenCalledWith([{
            type: '',
            token: '',
            isDefault: false,
            details: '{}'
        }]);
    });

    it('updates payment method type', () => {
        const onChange = vi.fn();
        const items = [{ type: 'CreditCard', token: '', isDefault: false }];
        render(<PaymentMethodsForm items={items} onChange={onChange} />);

        // Assuming the select/input has a label 'Type' or similar. 
        // Based on my previous form implementations, it likely has a label.
        // Let's check the implementation or use a safe bet like placeholder if label is missing
        // or just use getByDisplayValue as I did before but be careful.
        // Actually PaymentMethodsForm probably has a label "Type".

        const typeInput = screen.getByLabelText('Type');
        fireEvent.change(typeInput, { target: { value: 'BankTransfer' } });

        expect(onChange).toHaveBeenCalledWith([{
            type: 'BankTransfer',
            token: '',
            isDefault: false
        }]);
    });
});
