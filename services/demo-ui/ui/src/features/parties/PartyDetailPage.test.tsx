import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import PartyDetailPage from './PartyDetailPage';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import * as api from './api';
import { type Individual, type Organization } from './types';

vi.mock('./api');

const mockIndividual: Individual = {
    id: 'p1',
    '@type': 'Individual',
    givenName: 'John',
    familyName: 'Doe',
    status: 'active',
    contactMediums: [{ id: 'cm1', mediumType: 'email', value: 'john@example.com', preferred: true }],
    identifications: [],
    relatedParties: []
};

const mockOrganization: Organization = {
    id: 'p2',
    '@type': 'Organization',
    tradingName: 'Acme Corp',
    isLegalEntity: true,
    status: 'active',
    contactMediums: [],
    identifications: [],
    relatedParties: []
};

describe('PartyDetailPage', () => {
    beforeEach(() => {
        vi.resetAllMocks();
        vi.mocked(api.useDeleteParty).mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
            // Cast to satisfy type without full mock implementation
        } as unknown as ReturnType<typeof api.useDeleteParty>);
    });

    it('renders loading state', () => {
        vi.mocked(api.useParty).mockReturnValue({
            data: undefined,
            isLoading: true,
            error: null,
        } as unknown as ReturnType<typeof api.useParty>);

        render(
            <MemoryRouter initialEntries={['/parties/p1']}>
                <Routes>
                    <Route path="/parties/:id" element={<PartyDetailPage />} />
                </Routes>
            </MemoryRouter>
        );

        expect(screen.getByRole('status')).toHaveTextContent('Loading party details...');
    });

    it('renders individual details', () => {
        vi.mocked(api.useParty).mockReturnValue({
            data: mockIndividual,
            isLoading: false,
            error: null,
        } as unknown as ReturnType<typeof api.useParty>);

        render(
            <MemoryRouter initialEntries={['/parties/p1']}>
                <Routes>
                    <Route path="/parties/:id" element={<PartyDetailPage />} />
                </Routes>
            </MemoryRouter>
        );

        expect(screen.getByRole('heading', { name: 'John Doe' })).toBeInTheDocument();
        expect(screen.getAllByText('Individual')[0]).toBeVisible(); // Badge
        expect(screen.getByText('john@example.com')).toBeInTheDocument();
    });

    it('renders organization details', () => {
        vi.mocked(api.useParty).mockReturnValue({
            data: mockOrganization,
            isLoading: false,
            error: null,
        } as unknown as ReturnType<typeof api.useParty>);

        render(
            <MemoryRouter initialEntries={['/parties/p2']}>
                <Routes>
                    <Route path="/parties/:id" element={<PartyDetailPage />} />
                </Routes>
            </MemoryRouter>
        );

        expect(screen.getByRole('heading', { name: 'Acme Corp' })).toBeInTheDocument();
        expect(screen.getAllByText('Organization')[0]).toBeVisible();
        expect(screen.getByText('Yes')).toBeInTheDocument(); // Legal Entity
    });
});
