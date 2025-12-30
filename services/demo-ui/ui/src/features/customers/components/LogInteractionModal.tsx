import { useState } from 'react';
import { useLogInteraction } from '../api';
import { X, Loader2 } from 'lucide-react';
import './LogInteractionModal.css';

interface LogInteractionModalProps {
    customerId: string;
    onClose: () => void;
}

export default function LogInteractionModal({ customerId, onClose }: LogInteractionModalProps) {
    const logInteractionMutation = useLogInteraction();

    const [type, setType] = useState('Call');
    const [channel, setChannel] = useState('Phone');
    const [description, setDescription] = useState('');
    const [agentId, setAgentId] = useState('admin-1'); // Default logic needed

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        try {
            await logInteractionMutation.mutateAsync({
                id: crypto.randomUUID(),
                customerId,
                interactionDate: new Date().toISOString(),
                channel,
                type,
                description,
                agentId
            });
            onClose();
        } catch (error) {
            console.error('Failed to log interaction', error);
        }
    };

    return (
        <div className="modal-overlay">
            <div className="modal-content card">
                <div className="modal-header">
                    <h3>Log Interaction</h3>
                    <button type="button" className="btn-icon" onClick={onClose}>
                        <X size={20} />
                    </button>
                </div>
                <form onSubmit={handleSubmit}>
                    <div className="form-grid">
                        <div className="form-group">
                            <label htmlFor="int-type">Type</label>
                            <input
                                id="int-type"
                                type="text"
                                value={type}
                                onChange={(e) => setType(e.target.value)}
                                placeholder="e.g. Call, Email, Ticket"
                                required
                            />
                        </div>
                        <div className="form-group">
                            <label htmlFor="int-channel">Channel</label>
                            <input
                                id="int-channel"
                                type="text"
                                value={channel}
                                onChange={(e) => setChannel(e.target.value)}
                                placeholder="e.g. Phone, Web, App"
                                required
                            />
                        </div>
                        <div className="form-group">
                            <label htmlFor="int-agent">Agent ID</label>
                            <input
                                id="int-agent"
                                type="text"
                                value={agentId}
                                onChange={(e) => setAgentId(e.target.value)}
                                required
                            />
                        </div>
                        <div className="form-group form-group--full">
                            <label htmlFor="int-desc">Description</label>
                            <textarea
                                id="int-desc"
                                value={description}
                                onChange={(e) => setDescription(e.target.value)}
                                rows={4}
                                required
                            />
                        </div>
                    </div>

                    <div className="modal-actions">
                        <button type="button" className="btn btn-secondary" onClick={onClose}>Cancel</button>
                        <button type="submit" className="btn btn-primary" disabled={logInteractionMutation.isPending}>
                            {logInteractionMutation.isPending && <Loader2 className="spin" size={16} />}
                            Log Interaction
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
}
