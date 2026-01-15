import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import PriceEditor from './PriceEditor';
import type { ProductOfferingPrice } from '../types';

describe('PriceEditor', () => {
    const mockOnChange = vi.fn();
    const mockPrices: ProductOfferingPrice[] = [
        {
            priceType: 'one_time',
            price: { unit: 'EUR', value: 100 },
        },
        {
            priceType: 'recurring',
            price: { unit: 'EUR', value: 29.99 },
            unitOfMeasure: 'month'
        }
    ];

    it('renders empty state correctly', () => {
        render(<PriceEditor prices={[]} onChange={mockOnChange} />);
        expect(screen.getByText('No pricing defined for this offering.')).toBeInTheDocument();
        expect(screen.getByText('Pricing')).toBeInTheDocument();
        expect(screen.getByText('Add Price')).toBeInTheDocument();
    });

    it('renders existing prices correctly', () => {
        render(<PriceEditor prices={mockPrices} onChange={mockOnChange} />);

        const amountInputs = screen.getAllByRole('spinbutton');
        expect(amountInputs).toHaveLength(2);
        expect(amountInputs[0]).toHaveValue(100);
        expect(amountInputs[1]).toHaveValue(29.99);

        expect(screen.getByDisplayValue('month')).toBeInTheDocument();
    });

    it('adds a new price when Add Price is clicked', () => {
        render(<PriceEditor prices={[]} onChange={mockOnChange} />);

        fireEvent.click(screen.getByText('Add Price'));

        expect(mockOnChange).toHaveBeenCalledWith([{
            priceType: 'one_time',
            price: { unit: 'EUR', value: 0 }
        }]);
    });

    it('removes a price when delete button is clicked', () => {
        render(<PriceEditor prices={mockPrices} onChange={mockOnChange} />);

        const deleteButtons = screen.getAllByRole('button').filter(btn => btn.className.includes('icon-btn--danger'));
        fireEvent.click(deleteButtons[0]);

        expect(mockOnChange).toHaveBeenCalledWith([mockPrices[1]]);
    });

    it('updates price values correctly', () => {
        render(<PriceEditor prices={mockPrices} onChange={mockOnChange} />);

        const amountInputs = screen.getAllByRole('spinbutton');
        fireEvent.change(amountInputs[0], { target: { value: '150' } });

        const expected = [...mockPrices];
        expected[0] = { ...expected[0], price: { ...expected[0].price, value: 150 } };

        expect(mockOnChange).toHaveBeenCalledWith(expected);
    });

    it('shows unit of measure input only for recurring prices', () => {
        render(<PriceEditor prices={mockPrices} onChange={mockOnChange} />);

        // Should be visible for recurring
        expect(screen.getByDisplayValue('month')).toBeVisible();

        // Change recurring to one_time
        const selects = screen.getAllByRole('combobox');
        fireEvent.change(selects[1], { target: { value: 'one_time' } });

        const expected = [...mockPrices];
        expected[1] = { ...expected[1], priceType: 'one_time' };

        expect(mockOnChange).toHaveBeenCalledWith(expected);
    });
});
