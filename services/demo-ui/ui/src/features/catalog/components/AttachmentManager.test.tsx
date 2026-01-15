import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import AttachmentManager from './AttachmentManager';
import type { Attachment } from '../types';

describe('AttachmentManager', () => {
    const mockOnChange = vi.fn();
    const mockAttachments: Attachment[] = [
        {
            id: '1',
            name: 'brochure.pdf',
            url: 'http://example.com/brochure.pdf',
            type: 'Document',
            mimeType: 'application/pdf'
        }
    ];

    beforeEach(() => {
        vi.restoreAllMocks();
        mockOnChange.mockClear();
    });

    it('renders empty state correctly', () => {
        render(<AttachmentManager attachments={[]} onChange={mockOnChange} />);
        expect(screen.getByText('No attachments defined for this offering.')).toBeInTheDocument();
        expect(screen.getByText('Add Attachment')).toBeInTheDocument();
    });

    it('renders existing attachments correctly', () => {
        render(<AttachmentManager attachments={mockAttachments} onChange={mockOnChange} />);
        expect(screen.getByText('brochure.pdf')).toBeInTheDocument();
        expect(screen.getByText('http://example.com/brochure.pdf')).toBeInTheDocument();
    });

    it('adds a new attachment via prompt', () => {
        const promptSpy = vi.spyOn(window, 'prompt').mockReturnValue('http://test.com/new.pdf');

        // Mock crypto.randomUUID
        const mockUUID = 'new-uuid-123';
        vi.stubGlobal('crypto', { randomUUID: () => mockUUID });

        render(<AttachmentManager attachments={mockAttachments} onChange={mockOnChange} />);

        fireEvent.click(screen.getByText('Add Attachment'));

        expect(promptSpy).toHaveBeenCalled();
        expect(mockOnChange).toHaveBeenCalledWith([
            ...mockAttachments,
            {
                id: mockUUID,
                url: 'http://test.com/new.pdf',
                type: 'Document',
                mimeType: 'text/html',
                name: 'new.pdf'
            }
        ]);

        vi.unstubAllGlobals();
    });

    it('does not add attachment if prompt canceled', () => {
        const promptSpy = vi.spyOn(window, 'prompt').mockReturnValue(null);
        render(<AttachmentManager attachments={mockAttachments} onChange={mockOnChange} />);

        fireEvent.click(screen.getByText('Add Attachment'));

        expect(promptSpy).toHaveBeenCalled();
        expect(mockOnChange).not.toHaveBeenCalled();
    });

    it('removes an attachment', () => {
        render(<AttachmentManager attachments={mockAttachments} onChange={mockOnChange} />);

        // Or find by class since we know it
        const buttons = screen.getAllByRole('button');
        const deleteBtn = buttons.find(b => b.className.includes('btn-icon--danger'));

        fireEvent.click(deleteBtn!);

        expect(mockOnChange).toHaveBeenCalledWith([]);
    });
});
