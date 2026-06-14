import React from 'react';
import { FileText } from 'lucide-react';
import type { TaxExemption } from '../types';
import { formatDate } from '../../../lib/date';

interface TaxExemptionsSectionProps {
    exemptions?: TaxExemption[];
}

export const TaxExemptionsSection: React.FC<TaxExemptionsSectionProps> = ({ exemptions }) => {
    return (
        <div className="card detail-card">
            <h3>
                <span>Tax Exemptions</span>
                <span className="count-badge">{exemptions?.length || 0}</span>
            </h3>
            {exemptions && exemptions.length > 0 ? (
                <ul className="detail-list">
                    {exemptions.map((exemption) => (
                        <li key={exemption.id} className="detail-item-row">
                            <FileText size={18} className="text-muted" />
                            <div className="detail-content">
                                <span className="detail-label">{exemption.issuingJurisdiction}</span>
                                <span className="detail-value">{exemption.certificateNumber}</span>
                                <span className="detail-subtext">
                                    {exemption.validFor ? (
                                        <>
                                            Valid: {formatDate(exemption.validFor.startDateTime)}
                                            {exemption.validFor.endDateTime ? ` - ${formatDate(exemption.validFor.endDateTime)}` : ' (Indefinite)'}
                                        </>
                                    ) : (
                                        <span className="text-muted">No validity period</span>
                                    )}
                                </span>
                            </div>
                        </li>
                    ))}
                </ul>
            ) : (
                <p className="empty-text">No tax exemptions</p>
            )}
        </div>
    );
};
