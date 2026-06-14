import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { TaxExemptionsSection } from './TaxExemptionsSection';
import type { TaxExemption } from '../types';

const base: TaxExemption = {
    id: 't1',
    partyId: 'p1',
    certificateNumber: 'CERT-1',
    issuingJurisdiction: 'Estonia',
    validFor: { startDateTime: new Date(2026, 0, 31).toISOString() },
};

describe('TaxExemptionsSection', () => {
    it('shows an empty message with no exemptions', () => {
        render(<TaxExemptionsSection exemptions={[]} />);
        expect(screen.getByText('No tax exemptions')).toBeInTheDocument();
    });

    it('shows an open-ended validity as Indefinite with an Estonian start date', () => {
        render(<TaxExemptionsSection exemptions={[base]} />);
        expect(screen.getByText('Estonia')).toBeInTheDocument();
        expect(screen.getByText('CERT-1')).toBeInTheDocument();
        expect(screen.getByText(/Valid: 31\.01\.2026 \(Indefinite\)/)).toBeInTheDocument();
    });

    it('shows a start - end range when an end date is present', () => {
        const exemption: TaxExemption = {
            ...base,
            validFor: {
                startDateTime: new Date(2026, 0, 31).toISOString(),
                endDateTime: new Date(2026, 5, 14).toISOString(),
            },
        };
        render(<TaxExemptionsSection exemptions={[exemption]} />);
        expect(screen.getByText(/Valid: 31\.01\.2026 - 14\.06\.2026/)).toBeInTheDocument();
    });

    it('shows a no-validity message when validFor is absent', () => {
        const exemption = { ...base, validFor: undefined } as unknown as TaxExemption;
        render(<TaxExemptionsSection exemptions={[exemption]} />);
        expect(screen.getByText('No validity period')).toBeInTheDocument();
    });
});
