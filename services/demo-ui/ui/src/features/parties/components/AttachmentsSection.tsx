import React from 'react';
import { Paperclip, Download } from 'lucide-react';
import type { Attachment } from '../types';

interface AttachmentsSectionProps {
    attachments?: Attachment[];
}

export const AttachmentsSection: React.FC<AttachmentsSectionProps> = ({ attachments }) => {
    return (
        <div className="card detail-card">
            <h3>
                <span>Attachments</span>
                <span className="count-badge">{attachments?.length || 0}</span>
            </h3>
            {attachments && attachments.length > 0 ? (
                <ul className="detail-list">
                    {attachments.map((att) => (
                        <li key={att.id} className="detail-item-row">
                            <Paperclip size={18} className="text-muted" />
                            <div className="detail-content">
                                <span className="detail-label">{att.name}</span>
                                <span className="detail-subtext">{att.type} • {att.mimeType}</span>
                            </div>
                            {att.url && (
                                <a href={att.url} target="_blank" rel="noopener noreferrer" className="btn btn-icon">
                                    <Download size={16} />
                                </a>
                            )}
                        </li>
                    ))}
                </ul>
            ) : (
                <p className="empty-text">No attachments</p>
            )}
        </div>
    );
};
