import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import CharacteristicEditor from './CharacteristicEditor';
import type { ProductSpecCharacteristic } from '../types';

describe('CharacteristicEditor', () => {
    const mockOnChange = vi.fn();
    const mockCharacteristics: Record<string, ProductSpecCharacteristic> = {
        'Color': {
            name: 'Color',
            valueType: 'string',
            configurable: true,
            description: 'Product color'
        },
        'Size': {
            name: 'Size',
            valueType: 'number',
            configurable: false
        }
    };

    it('renders empty state correctly', () => {
        render(<CharacteristicEditor characteristics={{}} onChange={mockOnChange} />);
        expect(screen.getByText('No characteristics defined.')).toBeInTheDocument();
        expect(screen.getByText('Add Characteristic')).toBeInTheDocument();
    });

    it('renders existing characteristics correctly', () => {
        render(<CharacteristicEditor characteristics={mockCharacteristics} onChange={mockOnChange} />);

        expect(screen.getByDisplayValue('Color')).toBeInTheDocument();
        expect(screen.getByDisplayValue('Size')).toBeInTheDocument();

        // Check value types
        // Selected options are harder to query directly by text, check select values
        const selects = screen.getAllByRole('combobox');
        expect(selects[0]).toHaveValue('string');
        expect(selects[1]).toHaveValue('number');
    });

    it('adds a new characteristic', () => {
        render(<CharacteristicEditor characteristics={{}} onChange={mockOnChange} />);

        fireEvent.click(screen.getByText('Add Characteristic'));

        expect(mockOnChange).toHaveBeenCalledWith({
            'New Characteristic 1': {
                name: 'New Characteristic 1',
                valueType: 'string',
                configurable: true
            }
        });
    });

    it('removes a characteristic', () => {
        render(<CharacteristicEditor characteristics={mockCharacteristics} onChange={mockOnChange} />);

        const deleteButtons = screen.getAllByRole('button').filter(btn => btn.className.includes('btn-icon--danger'));
        fireEvent.click(deleteButtons[0]);

        const expected = { ...mockCharacteristics };
        delete expected['Color'];
        expect(mockOnChange).toHaveBeenCalledWith(expected);
    });

    it('updates characteristic name (key change)', () => {
        render(<CharacteristicEditor characteristics={mockCharacteristics} onChange={mockOnChange} />);

        const nameInput = screen.getByDisplayValue('Color');
        fireEvent.change(nameInput, { target: { value: 'Colour' } });

        const expected = { ...mockCharacteristics };
        delete expected['Color'];
        expected['Colour'] = { ...mockCharacteristics['Color'], name: 'Colour' };

        expect(mockOnChange).toHaveBeenCalledWith(expected);
    });

    it('updates characteristic properties', () => {
        render(<CharacteristicEditor characteristics={mockCharacteristics} onChange={mockOnChange} />);

        const checkbox = screen.getAllByRole('checkbox')[0]; // Color is configurable=true
        fireEvent.click(checkbox); // Toggle to false

        const expected = { ...mockCharacteristics };
        expected['Color'] = { ...expected['Color'], configurable: false };

        expect(mockOnChange).toHaveBeenCalledWith(expected);
    });
});
