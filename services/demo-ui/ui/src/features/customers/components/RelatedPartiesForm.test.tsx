import { render, screen, fireEvent } from '@testing-library/react';
import RelatedPartiesForm from './RelatedPartiesForm';
import { vi, describe, it, expect, beforeEach } from 'vitest';

// Mock hooks
vi.mock('../../parties/api');
vi.mock('@tanstack/react-query', async () => {
    const actual = await vi.importActual('@tanstack/react-query');
    return {
        ...actual,
        useQueryClient: vi.fn(() => ({
            invalidateQueries: vi.fn()
        })),
        useQuery: vi.fn(() => ({
            data: [],
            isLoading: false,
            error: null
        })),
    };
});

import * as api from '../../parties/api';

describe('RelatedPartiesForm', () => {
    beforeEach(() => {
        vi.resetAllMocks();
        // Mock default api response
        (api.useParties as import('vitest').Mock).mockReturnValue({
            data: [],
            isLoading: false,
            error: null,
            refetch: vi.fn(),
            isFetching: false
        });
    });

    it('renders empty state correctly', () => {
        render(<RelatedPartiesForm items={[]} onChange={vi.fn()} />);
        expect(screen.getByText('No related parties added')).toBeInTheDocument();
        expect(screen.getByText('Add Party')).toBeInTheDocument();
    });

    it('adds a new related party', () => {
        const onChange = vi.fn();
        render(<RelatedPartiesForm items={[]} onChange={onChange} />);

        fireEvent.click(screen.getByText('Add Party'));

        expect(onChange).toHaveBeenCalledWith([{
            relatedPartyId: '',
            name: '',
            role: ''
        }]);
    });

    it('removes a related party', () => {
        const onChange = vi.fn();
        const items = [{ relatedPartyId: '1', name: 'Test', role: 'Role' }];
        render(<RelatedPartiesForm items={items} onChange={onChange} />);

        fireEvent.click(screen.getByRole('button', { name: /remove/i }));

        expect(onChange).toHaveBeenCalledWith([]);
    });

    it('updates a related party field', () => {
        const onChange = vi.fn();
        const items = [{ relatedPartyId: '', name: '', role: '' }];
        render(<RelatedPartiesForm items={items} onChange={onChange} />);

        const roleInput = screen.getByLabelText('Role');
        fireEvent.change(roleInput, { target: { value: 'New Role' } });

        expect(onChange).toHaveBeenCalledWith([{
            relatedPartyId: '',
            name: '',
            role: 'New Role'
        }]);
    });
});
