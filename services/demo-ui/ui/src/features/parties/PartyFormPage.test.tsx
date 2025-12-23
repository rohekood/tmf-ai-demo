import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import PartyFormPage from './PartyFormPage';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import * as api from './api';

vi.mock('./api');

// Mock navigate
const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom');
    return {
        ...actual,
        useNavigate: () => mockNavigate,
    };
});

describe('PartyFormPage', () => {
    let createMock: any;
    let updateMock: any;

    beforeEach(() => {
        vi.resetAllMocks();
        createMock = vi.fn();
        updateMock = vi.fn();

        (api.useCreateParty as any).mockReturnValue({
            mutateAsync: createMock,
            isPending: false
        });
        (api.useUpdateParty as any).mockReturnValue({
            mutateAsync: updateMock,
            isPending: false
        });
    });

    it('renders create form (default Individual)', () => {
        (api.useParty as any).mockReturnValue({ data: undefined, isLoading: false });

        render(
            <MemoryRouter initialEntries={['/parties/new']}>
                <Routes>
                    <Route path="/parties/new" element={<PartyFormPage />} />
                </Routes>
            </MemoryRouter>
        );

        expect(screen.getByRole('heading', { name: 'Create Party' })).toBeInTheDocument();
        expect(screen.getByLabelText('Given Name *')).toBeInTheDocument();
    });

    it('toggles to Organization', async () => {
        const user = userEvent.setup();
        (api.useParty as any).mockReturnValue({ data: undefined, isLoading: false });

        render(
            <MemoryRouter initialEntries={['/parties/new']}>
                <Routes>
                    <Route path="/parties/new" element={<PartyFormPage />} />
                </Routes>
            </MemoryRouter>
        );

        const orgRadio = screen.getByRole('radio', { name: /Organization/i });
        await user.click(orgRadio);

        expect(screen.getByLabelText('Trading Name *')).toBeInTheDocument();
        expect(screen.queryByLabelText('Given Name *')).not.toBeInTheDocument();
    });

    it('submits create Individual', async () => {
        const user = userEvent.setup();
        (api.useParty as any).mockReturnValue({ data: undefined, isLoading: false });

        render(
            <MemoryRouter initialEntries={['/parties/new']}>
                <Routes>
                    <Route path="/parties/new" element={<PartyFormPage />} />
                </Routes>
            </MemoryRouter>
        );

        await user.type(screen.getByLabelText('Given Name *'), 'John');
        await user.type(screen.getByLabelText('Family Name *'), 'Doe');

        const submitBtn = screen.getByRole('button', { name: 'Create Party' });
        await user.click(submitBtn);

        await waitFor(() => {
            expect(createMock).toHaveBeenCalledWith(expect.objectContaining({
                '@type': 'Individual',
                givenName: 'John',
                familyName: 'Doe'
            }));
            expect(mockNavigate).toHaveBeenCalledWith('/parties');
        });
    });

    it('renders edit form', () => {
        const mockParty = {
            id: 'p1',
            '@type': 'Individual',
            givenName: 'John',
            familyName: 'Doe',
            contactMediums: [],
            identifications: []
        };
        (api.useParty as any).mockReturnValue({ data: mockParty, isLoading: false });

        render(
            <MemoryRouter initialEntries={['/parties/p1/edit']}>
                <Routes>
                    <Route path="/parties/:id/edit" element={<PartyFormPage />} />
                </Routes>
            </MemoryRouter>
        );

        expect(screen.getByDisplayValue('John')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Save Changes' })).toBeInTheDocument();
    });
});
