import { describe, it, expect } from 'vitest';
import { formatDate, formatDateTime } from './date';

describe('formatDate', () => {
    it('formats a Date as an Estonian date (dd.MM.yyyy)', () => {
        // Local constructor keeps the assertion timezone-independent.
        expect(formatDate(new Date(2026, 5, 14))).toBe('14.06.2026');
        expect(formatDate(new Date(2026, 0, 31))).toBe('31.01.2026');
    });

    it('formats a local ISO string', () => {
        expect(formatDate('2026-06-14T12:00:00')).toBe('14.06.2026');
    });

    it('returns the fallback for empty/invalid input', () => {
        expect(formatDate(undefined)).toBe('');
        expect(formatDate(null)).toBe('');
        expect(formatDate('')).toBe('');
        expect(formatDate('not-a-date')).toBe('');
        expect(formatDate(undefined, 'Start')).toBe('Start');
    });
});

describe('formatDateTime', () => {
    it('formats a Date as an Estonian date-time (dd.MM.yyyy HH:mm:ss, no comma)', () => {
        expect(formatDateTime(new Date(2026, 5, 14, 9, 25, 4))).toBe('14.06.2026 09:25:04');
    });

    it('returns the fallback for missing input', () => {
        expect(formatDateTime(null)).toBe('');
        expect(formatDateTime(undefined, '—')).toBe('—');
    });
});
