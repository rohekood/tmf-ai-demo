import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { Layout } from './Layout';
import { MemoryRouter } from 'react-router-dom';

describe('Layout', () => {
    it('renders the layout with sidebar and headers', () => {
        render(
            <MemoryRouter>
                <Layout />
            </MemoryRouter>
        );

        // Check header
        expect(screen.getByRole('banner')).toBeInTheDocument();
        expect(screen.getByRole('heading', { name: 'TMF Demo Dashboard' })).toBeInTheDocument();
        // Assuming the subtitle isn't a heading, maybe text is fine or role="note"
        expect(screen.getByText('Managed via Golang BFF & RabbitMQ')).toBeInTheDocument();

        // Check Sidebar presence
        expect(screen.getByRole('navigation')).toBeInTheDocument();
    });
});
