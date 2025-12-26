import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import CustomerEditPage from './CustomerEditPage';
import { MemoryRouter, Route, Routes } from 'react-router-dom';

// Mocks
const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom');
    return {
        ...actual,
        useNavigate: () => mockNavigate,
    };
});

const mockCustomerData = {
    id: '123',
    name: 'Test Customer',
    status: 'active',
    taxExemptions: [],
    privacyConsents: []
};

const mockMutateAsync = vi.fn();
vi.mock('./api', () => ({
    useCustomer: (id: string) => ({
        data: id === '123' ? mockCustomerData : null,
        isLoading: false,
    }),
    useUpdateCustomer: () => ({
        mutateAsync: mockMutateAsync,
        isPending: false,
    }),
}));

describe('CustomerEditPage', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('renders customer data correctly', () => {
        render(
            <MemoryRouter initialEntries={['/customers/123/edit']}>
                <Routes>
                    <Route path="/customers/:id/edit" element={<CustomerEditPage />} />
                </Routes>
            </MemoryRouter>
        );

        expect(screen.getByDisplayValue('Test Customer')).toBeInTheDocument();
        expect(screen.getByDisplayValue('Active')).toBeInTheDocument();
    });

    it('submits form with updated data', async () => {
        render(
            <MemoryRouter initialEntries={['/customers/123/edit']}>
                <Routes>
                    <Route path="/customers/:id/edit" element={<CustomerEditPage />} />
                </Routes>
            </MemoryRouter>
        );

        const nameInput = screen.getByLabelText('Customer Name');
        fireEvent.change(nameInput, { target: { value: 'Updated Name' } });

        const saveButton = screen.getByRole('button', { name: /save changes/i });
        fireEvent.click(saveButton);

        await waitFor(() => {
            expect(mockMutateAsync).toHaveBeenCalledWith(expect.objectContaining({
                id: '123',
                name: 'Updated Name',
            }));
        });

        expect(mockNavigate).toHaveBeenCalledWith('/customers/123');
    });

    it('adds a tax exemption', () => {
        render(
            <MemoryRouter initialEntries={['/customers/123/edit']}>
                <Routes>
                    <Route path="/customers/:id/edit" element={<CustomerEditPage />} />
                </Routes>
            </MemoryRouter>
        );

        const addButton = screen.getByRole('button', { name: /add exemption/i });
        fireEvent.click(addButton);

        expect(screen.getByPlaceholderText('Certificate #')).toBeInTheDocument();
        expect(screen.getByPlaceholderText('e.g., US-CA')).toBeInTheDocument();
    });
});
