import { Plus, Trash2, Paperclip, Link as LinkIcon } from 'lucide-react';
import type { Attachment } from '../types';

interface AttachmentManagerProps {
    attachments: Attachment[];
    onChange: (attachments: Attachment[]) => void;
}

export default function AttachmentManager({ attachments, onChange }: AttachmentManagerProps) {
    const handleAdd = () => {
        const url = prompt('Enter attachment URL:');
        if (url) {
            onChange([
                ...attachments,
                {
                    id: crypto.randomUUID(),
                    url,
                    type: 'Document',
                    mimeType: 'text/html',
                    name: url.split('/').pop() || 'Untitled',
                }
            ]);
        }
    };

    const handleRemove = (index: number) => {
        onChange(attachments.filter((_, i) => i !== index));
    };

    return (
        <div className="card form-section">
            <div className="section-header">
                <h3>Attachments</h3>
                <button type="button" className="btn btn-secondary btn-sm" onClick={handleAdd}>
                    <Plus size={16} />
                    <span>Add Attachment</span>
                </button>
            </div>

            {attachments.length === 0 ? (
                <p className="empty-text">No attachments defined for this offering.</p>
            ) : (
                <div className="repeatable-list">
                    {attachments.map((a, index) => (
                        <div key={index} className="repeatable-item">
                            <div className="row-between">
                                <div className="row" style={{ minWidth: 0 }}>
                                    <Paperclip size={16} className="muted" />
                                    <div className="stack-sm" style={{ gap: '0.15rem', minWidth: 0 }}>
                                        <span style={{ fontWeight: 600 }}>{a.name}</span>
                                        <a href={a.url} target="_blank" rel="noopener noreferrer" className="row" style={{ gap: '0.35rem', fontSize: '0.8125rem' }}>
                                            <LinkIcon size={12} />
                                            {a.url}
                                        </a>
                                    </div>
                                </div>
                                <button
                                    type="button"
                                    className="btn-icon btn-icon--danger"
                                    onClick={() => handleRemove(index)}
                                >
                                    <Trash2 size={16} />
                                </button>
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}
