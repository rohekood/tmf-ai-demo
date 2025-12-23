import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { Sidebar } from './Sidebar';
import { MemoryRouter } from 'react-router-dom';

describe('Sidebar', () => {
    it('renders navigation links', () => {
        render(
            <MemoryRouter>
                <Sidebar />
            </MemoryRouter>
        );

        expect(screen.getByRole('link', { name: 'Parties' })).toBeInTheDocument();
        expect(screen.getByRole('link', { name: 'Customers' })).toBeInTheDocument();
        expect(screen.getByRole('link', { name: 'Debug Console' })).toBeInTheDocument();
        expect(screen.getByRole('status')).toHaveTextContent('Connected');
    });
});
