import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import PartyDetailPage from './PartyDetailPage';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import * as api from './api';

vi.mock('./api');

const mockIndividual = {
    id: 'p1',
    '@type': 'Individual',
    givenName: 'John',
    familyName: 'Doe',
    status: 'Active',
    contactMediums: [{ id: 'cm1', mediumType: 'email', value: 'john@example.com' }],
    identifications: [],
    relatedParties: []
};

const mockOrganization = {
    id: 'p2',
    '@type': 'Organization',
    tradingName: 'Acme Corp',
    isLegalEntity: true,
    status: 'Active',
    contactMediums: [],
    identifications: [],
    relatedParties: []
};

describe('PartyDetailPage', () => {
    beforeEach(() => {
        vi.resetAllMocks();
        (api.useDeleteParty as any).mockReturnValue({
            mutate: vi.fn(),
            isPending: false
        });
    });

    it('renders loading state', () => {
        (api.useParty as any).mockReturnValue({
            data: undefined,
            isLoading: true,
            error: null
        });

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
        (api.useParty as any).mockReturnValue({
            data: mockIndividual,
            isLoading: false,
            error: null
        });

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
        (api.useParty as any).mockReturnValue({
            data: mockOrganization,
            isLoading: false,
            error: null
        });

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
