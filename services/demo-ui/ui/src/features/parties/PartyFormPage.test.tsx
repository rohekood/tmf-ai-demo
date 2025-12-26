import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import PartyFormPage from './PartyFormPage';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import * as api from './api';
import { type Individual } from './types';

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
    let createMock: ReturnType<typeof vi.fn>;
    let updateMock: ReturnType<typeof vi.fn>;

    beforeEach(() => {
        vi.resetAllMocks();
        createMock = vi.fn();
        updateMock = vi.fn();

        vi.mocked(api.useCreateParty).mockReturnValue({
            mutateAsync: createMock,
            isPending: false
        } as unknown as ReturnType<typeof api.useCreateParty>);
        vi.mocked(api.useUpdateParty).mockReturnValue({
            mutateAsync: updateMock,
            isPending: false
        } as unknown as ReturnType<typeof api.useUpdateParty>);
    });

    it('renders create form (default Individual)', () => {
        vi.mocked(api.useParty).mockReturnValue({ data: undefined, isLoading: false } as unknown as ReturnType<typeof api.useParty>);

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
        vi.mocked(api.useParty).mockReturnValue({ data: undefined, isLoading: false } as unknown as ReturnType<typeof api.useParty>);

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
        vi.mocked(api.useParty).mockReturnValue({ data: undefined, isLoading: false } as unknown as ReturnType<typeof api.useParty>);

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
        const mockParty: Individual = {
            id: 'p1',
            '@type': 'Individual',
            givenName: 'John',
            familyName: 'Doe',
            status: 'active', // Ensure valid status
            contactMediums: [],
            identifications: []
        };
        vi.mocked(api.useParty).mockReturnValue({ data: mockParty, isLoading: false } as unknown as ReturnType<typeof api.useParty>);

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
    it('populates form after async data load', async () => {
        const mockParty: Individual = {
            id: 'p2',
            '@type': 'Individual',
            givenName: 'Jane',
            familyName: 'Smith',
            status: 'active',
            contactMediums: [],
            identifications: []
        };

        // First render with loading state
        vi.mocked(api.useParty).mockReturnValue({
            data: undefined,
            isLoading: true
        } as unknown as ReturnType<typeof api.useParty>);

        const { rerender } = render(
            <MemoryRouter initialEntries={['/parties/p2/edit']}>
                <Routes>
                    <Route path="/parties/:id/edit" element={<PartyFormPage />} />
                </Routes>
            </MemoryRouter>
        );

        expect(screen.getByText('Loading party...')).toBeInTheDocument();

        // Second render with data
        vi.mocked(api.useParty).mockReturnValue({
            data: mockParty,
            isLoading: false
        } as unknown as ReturnType<typeof api.useParty>);

        rerender(
            <MemoryRouter initialEntries={['/parties/p2/edit']}>
                <Routes>
                    <Route path="/parties/:id/edit" element={<PartyFormPage />} />
                </Routes>
            </MemoryRouter>
        );

        await waitFor(() => {
            expect(screen.getByDisplayValue('Jane')).toBeInTheDocument();
            expect(screen.getByDisplayValue('Smith')).toBeInTheDocument();
        });
    });
    it('submits update with correct flat payload and all fields', async () => {
        const user = userEvent.setup();
        const mockParty: Individual = {
            id: 'p1',
            '@type': 'Individual',
            givenName: 'John',
            familyName: 'Doe',
            middleName: 'M',
            birthDate: '1990-01-01',
            gender: 'male',
            status: 'active',
            contactMediums: [{ id: 'c1', mediumType: 'email', preferred: true, value: 'john@example.com' }],
            identifications: []
        };

        vi.mocked(api.useParty).mockReturnValue({
            data: mockParty,
            isLoading: false
        } as unknown as ReturnType<typeof api.useParty>);

        render(
            <MemoryRouter initialEntries={['/parties/p1/edit']}>
                <Routes>
                    <Route path="/parties/:id/edit" element={<PartyFormPage />} />
                </Routes>
            </MemoryRouter>
        );

        // Modify fields
        const givenNameInput = screen.getByLabelText('Given Name *');
        await user.clear(givenNameInput);
        await user.type(givenNameInput, 'Johnny');

        const middleNameInput = screen.getByLabelText('Middle Name');
        await user.clear(middleNameInput);
        await user.type(middleNameInput, 'X');

        const submitBtn = screen.getByRole('button', { name: 'Save Changes' });
        await user.click(submitBtn);

        await waitFor(() => {
            // Expect FLAT payload structure, NOT nested
            expect(updateMock).toHaveBeenCalledWith(expect.objectContaining({
                id: 'p1',
                '@type': 'Individual',
                givenName: 'Johnny',
                familyName: 'Doe',
                middleName: 'X',
                birthDate: '1990-01-01',
                gender: 'male',
                contactMediums: expect.arrayContaining([
                    expect.objectContaining({
                        mediumType: 'email',
                        value: 'john@example.com'
                    })
                ])
            }));
        });
    });

    it('submits create Organization with organizationType', async () => {
        const user = userEvent.setup();
        vi.mocked(api.useParty).mockReturnValue({ data: undefined, isLoading: false } as unknown as ReturnType<typeof api.useParty>);

        render(
            <MemoryRouter initialEntries={['/parties/new']}>
                <Routes>
                    <Route path="/parties/new" element={<PartyFormPage />} />
                </Routes>
            </MemoryRouter>
        );

        // Switch to Organization
        await user.click(screen.getByRole('radio', { name: /Organization/i }));

        await user.type(screen.getByLabelText('Trading Name *'), 'Acme Corp');
        await user.type(screen.getByPlaceholderText('e.g., LLC, Inc, GmbH'), 'LLC'); // organizationType input

        const submitBtn = screen.getByRole('button', { name: 'Create Party' });
        await user.click(submitBtn);

        await waitFor(() => {
            expect(createMock).toHaveBeenCalledWith(expect.objectContaining({
                '@type': 'Organization',
                tradingName: 'Acme Corp',
                organizationType: 'LLC',
                isLegalEntity: true // default
            }));
        });
    });

    it('switches type from Individual to Organization in edit mode', async () => {
        const user = userEvent.setup();
        const mockParty: Individual = {
            id: 'p3',
            '@type': 'Individual',
            givenName: 'John',
            familyName: 'Doe',
            status: 'active',
            contactMediums: [],
            identifications: []
        };

        vi.mocked(api.useParty).mockReturnValue({
            data: mockParty,
            isLoading: false
        } as unknown as ReturnType<typeof api.useParty>);

        render(
            <MemoryRouter initialEntries={['/parties/p3/edit']}>
                <Routes>
                    <Route path="/parties/:id/edit" element={<PartyFormPage />} />
                </Routes>
            </MemoryRouter>
        );

        // Switch to Organization
        const orgRadio = screen.getByRole('radio', { name: /Organization/i });
        // Check if enabled (user requirement)
        expect(orgRadio).not.toBeDisabled();
        await user.click(orgRadio);

        // Verify fields change
        expect(screen.getByLabelText('Trading Name *')).toBeInTheDocument();
        expect(screen.queryByLabelText('Given Name *')).not.toBeInTheDocument();

        // Fill Organization fields
        await user.type(screen.getByLabelText('Trading Name *'), 'Now A Corp');

        const submitBtn = screen.getByRole('button', { name: 'Save Changes' });
        await user.click(submitBtn);

        await waitFor(() => {
            expect(updateMock).toHaveBeenCalledWith(expect.objectContaining({
                id: 'p3',
                '@type': 'Organization', // Type changed
                tradingName: 'Now A Corp'
            }));
        });
    });
});
