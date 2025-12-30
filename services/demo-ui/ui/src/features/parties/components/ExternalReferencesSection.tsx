import React from 'react';
import { Globe } from 'lucide-react';
import type { ExternalReference } from '../types';

interface ExternalReferencesSectionProps {
    references?: ExternalReference[];
}

export const ExternalReferencesSection: React.FC<ExternalReferencesSectionProps> = ({ references }) => {
    return (
        <div className="card detail-card">
            <h3>
                <span>External References</span>
                <span className="count-badge">{references?.length || 0}</span>
            </h3>
            {references && references.length > 0 ? (
                <ul className="detail-list">
                    {references.map((ref) => (
                        <li key={ref.id} className="detail-item-row">
                            <Globe size={18} className="text-muted" />
                            <div className="detail-content">
                                <span className="detail-label">{ref.externalSystemId}</span>
                                <span className="detail-value">{ref.externalReference}</span>
                            </div>
                        </li>
                    ))}
                </ul>
            ) : (
                <p className="empty-text">No external references</p>
            )}
        </div>
    );
};
