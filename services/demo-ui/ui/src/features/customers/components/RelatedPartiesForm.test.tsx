import { render, screen, fireEvent } from '@testing-library/react';
import RelatedPartiesForm from './RelatedPartiesForm';
import { vi, describe, it, expect } from 'vitest';

describe('RelatedPartiesForm', () => {
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

        const nameInput = screen.getByLabelText('Party Name');
        fireEvent.change(nameInput, { target: { value: 'New Name' } });

        expect(onChange).toHaveBeenCalledWith([{
            relatedPartyId: '',
            name: 'New Name',
            role: ''
        }]);
    });
});
