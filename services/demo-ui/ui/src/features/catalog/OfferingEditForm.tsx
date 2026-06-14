import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Save, Loader2 } from 'lucide-react';
import { useCreateOffering, useUpdateOffering, useSpecifications } from './api';
import type { ProductOffering, LifecycleStatus, ProductOfferingPrice, CreateProductOfferingPayload } from './types';
import PriceEditor from './components/PriceEditor';
import CategoryPicker from './components/CategoryPicker';
import AttachmentManager from './components/AttachmentManager';
import { DateInput } from '../../design-system/components/common/DateInput';
import '../parties/PartyFormPage.css';

interface OfferingEditFormProps {
    offering?: ProductOffering;
    isNew: boolean;
}

export default function OfferingEditForm({ offering, isNew }: OfferingEditFormProps) {
    const navigate = useNavigate();
    const createMutation = useCreateOffering();
    const updateMutation = useUpdateOffering();

    const [name, setName] = useState(offering?.name || '');
    const [description, setDescription] = useState(offering?.description || '');
    const [lifecycleStatus, setLifecycleStatus] = useState<LifecycleStatus>(offering?.lifecycleStatus || 'Draft');
    const [isBundle, setIsBundle] = useState(offering?.isBundle || false);
    const [isSellable, setIsSellable] = useState(offering?.isSellable !== false);

    const [specId, setSpecId] = useState(offering?.productSpecificationId || '');
    const [prices, setPrices] = useState<ProductOfferingPrice[]>(offering?.productOfferingPrice || []);
    const [categoryIds, setCategoryIds] = useState<string[]>(offering?.categoryIds || []);
    const [attachments, setAttachments] = useState(offering?.attachments || []);

    const [startDate, setStartDate] = useState(offering?.validFor?.startDateTime?.split('T')[0] || '');
    const [endDate, setEndDate] = useState(offering?.validFor?.endDateTime?.split('T')[0] || '');

    const { data: specifications = [] } = useSpecifications();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        const payload = {
            name,
            description,
            lifecycleStatus,
            isBundle,
            isSellable,
            productSpecificationId: specId,
            productOfferingPrice: prices,
            categoryIds,
            attachments,
            validFor: {
                startDateTime: startDate ? `${startDate}T00:00:00Z` : undefined,
                endDateTime: endDate ? `${endDate}T23:59:59Z` : undefined,
            },
        };

        try {
            if (isNew) {
                const result = await createMutation.mutateAsync(payload as CreateProductOfferingPayload);
                navigate(`/catalog/offerings/${result.id}`);
            } else if (offering) {
                await updateMutation.mutateAsync({
                    id: offering.id,
                    payload: payload as Partial<ProductOffering>,
                });
                navigate(`/catalog/offerings/${offering.id}`);
            }
        } catch (err) {
            console.error('Failed to save offering:', err);
        }
    };

    const isPending = createMutation.isPending || updateMutation.isPending;

    return (
        <>
            <div className="page-header">
                <div className="page-header-content mb-3">
                    <h2>{isNew ? 'New Product Offering' : 'Edit Offering'}</h2>
                    <p className="lead text-muted mb-0">
                        {isNew ? 'Create a commercial offering' : `Editing ${offering?.name}`}
                    </p>
                </div>
                <button
                    type="submit"
                    form="offering-form"
                    className="btn btn-primary"
                    disabled={isPending}
                >
                    {isPending ? <Loader2 className="spin" size={18} /> : <Save size={18} />}
                    <span>{isNew ? 'Create Offering' : 'Save Changes'}</span>
                </button>
            </div>

            <form id="offering-form" className="form-container" onSubmit={handleSubmit}>
                <div className="card form-section">
                    <h3>Basic Information</h3>
                    <div className="form-grid">
                        <div className="form-group">
                            <label htmlFor="name">Offering Name</label>
                            <input
                                id="name"
                                type="text"
                                value={name}
                                onChange={(e) => setName(e.target.value)}
                                required
                                placeholder="e.g., iPhone 15 Pro - Premium Plan"
                            />
                        </div>
                        <div className="form-group">
                            <label htmlFor="spec">Product Specification</label>
                            <select
                                id="spec"
                                value={specId}
                                onChange={(e) => setSpecId(e.target.value)}
                                required
                            >
                                <option value="">Select a specification...</option>
                                {specifications.map(s => (
                                    <option key={s.id} value={s.id}>{s.name} ({s.productNumber})</option>
                                ))}
                            </select>
                        </div>
                        <div className="form-group">
                            <label htmlFor="status">Lifecycle Status</label>
                            <select
                                id="status"
                                value={lifecycleStatus}
                                onChange={(e) => setLifecycleStatus(e.target.value as LifecycleStatus)}
                            >
                                <option value="Draft">Draft</option>
                                <option value="Active">Active</option>
                                <option value="Retired">Retired</option>
                                <option value="Suspended">Suspended</option>
                            </select>
                        </div>
                        <div className="form-group" style={{ display: 'flex', gap: '1.5rem', marginTop: '1.5rem' }}>
                            <label className="checkbox-label">
                                <input
                                    type="checkbox"
                                    checked={isSellable}
                                    onChange={(e) => setIsSellable(e.target.checked)}
                                />
                                <span>Sellable</span>
                            </label>
                            <label className="checkbox-label">
                                <input
                                    type="checkbox"
                                    checked={isBundle}
                                    onChange={(e) => setIsBundle(e.target.checked)}
                                />
                                <span>Bundle</span>
                            </label>
                        </div>
                    </div>
                    <div className="form-group mt-3">
                        <label htmlFor="description">Description</label>
                        <textarea
                            id="description"
                            value={description}
                            onChange={(e) => setDescription(e.target.value)}
                            rows={3}
                            placeholder="Detailed commercial description..."
                        />
                    </div>
                </div>

                <div className="card form-section">
                    <h3>Categories</h3>
                    <CategoryPicker selectedIds={categoryIds} onChange={setCategoryIds} />
                </div>

                <PriceEditor prices={prices} onChange={setPrices} />

                <AttachmentManager attachments={attachments} onChange={setAttachments} />

                <div className="card form-section">
                    <h3>Validity Period</h3>
                    <div className="form-grid">
                        <div className="form-group">
                            <label htmlFor="startDate">Start Date</label>
                            <DateInput
                                id="startDate"
                                value={startDate}
                                onChange={(e) => setStartDate(e.target.value)}
                            />
                        </div>
                        <div className="form-group">
                            <label htmlFor="endDate">End Date</label>
                            <DateInput
                                id="endDate"
                                value={endDate}
                                onChange={(e) => setEndDate(e.target.value)}
                            />
                        </div>
                    </div>
                </div>
            </form>
        </>
    );
}
