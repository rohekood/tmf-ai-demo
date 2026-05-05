
import { render, screen, fireEvent, cleanup } from '@testing-library/react';
import { DebugFilters } from './DebugFilters';
import { describe, it, expect, vi } from 'vitest';
import type { DebugFilterState } from '../types';

describe('DebugFilters', () => {
    const defaultFilter: DebugFilterState = {
        search: '',
        services: [],
        types: [],
    };
    const mockOnChange = vi.fn();
    const mockOnClear = vi.fn();

    it('renders filters correctly', () => {
        render(
            <DebugFilters
                filter={defaultFilter}
                onChange={mockOnChange}
                onClear={mockOnClear}
                totalCount={100}
                filteredCount={50}
            />
        );

        expect(screen.getByText('Filters')).toBeInTheDocument();
        expect(screen.getByPlaceholderText('Search payload or topic...')).toBeInTheDocument();
        expect(screen.getByText('50 / 100')).toBeInTheDocument();
    });

    it('handles search input change', () => {
        render(
            <DebugFilters
                filter={defaultFilter}
                onChange={mockOnChange}
                onClear={mockOnClear}
                totalCount={0}
                filteredCount={0}
            />
        );

        const input = screen.getByPlaceholderText('Search payload or topic...');
        fireEvent.change(input, { target: { value: 'test' } });

        expect(mockOnChange).toHaveBeenCalledWith({ ...defaultFilter, search: 'test' });
    });

    it('toggles service filters correctly', () => {
        render(
            <DebugFilters
                filter={defaultFilter}
                onChange={mockOnChange}
                onClear={mockOnClear}
                totalCount={0}
                filteredCount={0}
            />
        );

        const partyChip = screen.getByText('party');
        fireEvent.click(partyChip);

        expect(mockOnChange).toHaveBeenCalledWith({ ...defaultFilter, services: ['party'] });

        // Test untoggle
        // Re-render with the 'party' service selected
        const activeFilter = { ...defaultFilter, services: ['party'] };
        cleanup(); // Clear previous render
        render(
            <DebugFilters
                filter={activeFilter}
                onChange={mockOnChange}
                onClear={mockOnClear}
                totalCount={0}
                filteredCount={0}
            />
        );

        const activeChip = screen.getByText('party');
        // Verify it looks active
        expect(activeChip).toHaveClass('active');

        fireEvent.click(activeChip);
        // Should return empty list now
        expect(mockOnChange).toHaveBeenCalledWith({ ...activeFilter, services: [] });
    });

    it('toggles type filters correctly', () => {
        render(
            <DebugFilters
                filter={defaultFilter}
                onChange={mockOnChange}
                onClear={mockOnClear}
                totalCount={0}
                filteredCount={0}
            />
        );

        const cmdChip = screen.getByText('command');
        fireEvent.click(cmdChip);
        expect(mockOnChange).toHaveBeenCalledWith({ ...defaultFilter, types: ['command'] });
    });

    it('renders ordering service filter chip', () => {
        render(
            <DebugFilters
                filter={defaultFilter}
                onChange={mockOnChange}
                onClear={mockOnClear}
                totalCount={0}
                filteredCount={0}
            />
        );

        expect(screen.getByText('ordering')).toBeInTheDocument();
    });

    it('toggles ordering service filter correctly', () => {
        render(
            <DebugFilters
                filter={defaultFilter}
                onChange={mockOnChange}
                onClear={mockOnClear}
                totalCount={0}
                filteredCount={0}
            />
        );

        fireEvent.click(screen.getByText('ordering'));
        expect(mockOnChange).toHaveBeenCalledWith({ ...defaultFilter, services: ['ordering'] });
    });

    it('calls onClear when clear button is clicked', () => {
        render(
            <DebugFilters
                filter={defaultFilter}
                onChange={mockOnChange}
                onClear={mockOnClear}
                totalCount={0}
                filteredCount={0}
            />
        );

        fireEvent.click(screen.getByText('Clear Messages'));
        expect(mockOnClear).toHaveBeenCalled();
    });
});
