/**
 * Centralised date formatting helpers.
 *
 * The whole UI displays dates in the Estonian locale (`et-EE`), which renders
 * dates as `dd.MM.yyyy` and date-times as `dd.MM.yyyy HH:mm:ss` (24-hour).
 * Always format dates through these helpers so the format stays consistent.
 */

export const DATE_LOCALE = 'et-EE';

type DateInput = string | number | Date | null | undefined;

// Explicit options force zero-padded day/month and a 24-hour clock, so the
// output is `dd.MM.yyyy` regardless of the runtime's ICU defaults (Node's CLDR
// for `et` is non-padded; browsers pad — explicit options make both agree).
const dateFormatter = new Intl.DateTimeFormat(DATE_LOCALE, {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
});

const dateTimeFormatter = new Intl.DateTimeFormat(DATE_LOCALE, {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
});

function toDate(value: DateInput): Date | null {
    if (value === null || value === undefined || value === '') return null;
    const date = value instanceof Date ? value : new Date(value);
    return Number.isNaN(date.getTime()) ? null : date;
}

/**
 * Format a value as an Estonian date: `dd.MM.yyyy` (e.g. `14.06.2026`).
 * Returns `fallback` (default empty string) for missing/invalid input.
 */
export function formatDate(value: DateInput, fallback = ''): string {
    const date = toDate(value);
    return date ? dateFormatter.format(date) : fallback;
}

/**
 * Format a value as an Estonian date-time: `dd.MM.yyyy HH:mm:ss`.
 * The locale's literal comma separator (`14.06.2026, 12:25:04`) is dropped so
 * the date and time are separated by a single space.
 * Returns `fallback` (default empty string) for missing/invalid input.
 */
export function formatDateTime(value: DateInput, fallback = ''): string {
    const date = toDate(value);
    if (!date) return fallback;
    return dateTimeFormatter
        .formatToParts(date)
        .map((part) => (part.type === 'literal' ? part.value.replace(/,\s*/g, ' ') : part.value))
        .join('');
}
