import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import CategoryPicker from './CategoryPicker';
import type { Category } from '../types';

// Mock dependencies
const mockCategories: Category[] = [
    { id: '1', name: 'Electronics', isRoot: true, lifecycleStatus: 'Active', validFor: {}, lastUpdate: '' },
    { id: '2', name: 'Smartphones', isRoot: false, parentId: '1', lifecycleStatus: 'Active', validFor: {}, lastUpdate: '' },
    { id: '3', name: 'Laptops', isRoot: false, parentId: '1', lifecycleStatus: 'Active', validFor: {}, lastUpdate: '' }
];

vi.mock('../api', () => ({
    useCategories: () => ({
        data: mockCategories,
        isLoading: false
    })
}));

describe('CategoryPicker', () => {
    const mockOnChange = vi.fn();

    it('renders selected categories correctly', () => {
        render(<CategoryPicker selectedIds={['2']} onChange={mockOnChange} />);

        // Should show selected tag
        expect(screen.getByText('Smartphones')).toBeInTheDocument();
        // Should show add button
        expect(screen.getByText('Add to Category')).toBeInTheDocument();
    });

    it('removes a category when X is clicked', () => {
        render(<CategoryPicker selectedIds={['2', '3']} onChange={mockOnChange} />);

        // Find remove buttons
        const removeButton = screen.getByRole('button', { name: 'Remove Smartphones' });
        fireEvent.click(removeButton);

        expect(mockOnChange).toHaveBeenCalledWith(['3']);
    });

    it('shows dropdown when Add Category is clicked', () => {
        render(<CategoryPicker selectedIds={[]} onChange={mockOnChange} />);

        fireEvent.click(screen.getByText('Add to Category'));

        // Search input should appear
        expect(screen.getByPlaceholderText('Search categories...')).toBeInTheDocument();
        // List items should appear
        expect(screen.getByText('Electronics')).toBeInTheDocument();
        expect(screen.getByText('Smartphones')).toBeInTheDocument();
    });

    it('selects a category from dropdown', () => {
        render(<CategoryPicker selectedIds={[]} onChange={mockOnChange} />);

        fireEvent.click(screen.getByText('Add to Category'));
        fireEvent.click(screen.getByText('Electronics'));

        expect(mockOnChange).toHaveBeenCalledWith(['1']);
    });

    it('filters categories in dropdown', () => {
        render(<CategoryPicker selectedIds={[]} onChange={mockOnChange} />);

        fireEvent.click(screen.getByText('Add to Category'));
        const searchInput = screen.getByPlaceholderText('Search categories...');

        fireEvent.change(searchInput, { target: { value: 'Lap' } });

        expect(screen.getByText('Laptops')).toBeInTheDocument();
        expect(screen.queryByText('Smartphones')).not.toBeInTheDocument();
    });

    it('does not show already selected categories in dropdown', () => {
        render(<CategoryPicker selectedIds={['1']} onChange={mockOnChange} />);

        fireEvent.click(screen.getByText('Add to Category'));


        const dropdownOption = screen.queryByRole('button', {
            name: (content, element) => {
                return element.className === 'picker-item' && content.includes('Electronics');
            }
        });

        expect(dropdownOption).not.toBeInTheDocument();
        expect(screen.getByText('Smartphones')).toBeInTheDocument();
    });
});
