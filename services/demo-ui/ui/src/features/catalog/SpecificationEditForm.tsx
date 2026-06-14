import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Save, Loader2 } from 'lucide-react';
import { useCreateSpecification, useUpdateSpecification } from './api';
import type { ProductSpecification, LifecycleStatus, ProductSpecCharacteristic, CreateProductSpecificationPayload } from './types';
import CharacteristicEditor from './components/CharacteristicEditor';
import { DateInput } from '../../design-system/components/common/DateInput';
import '../parties/PartyFormPage.css';

interface SpecificationEditFormProps {
    specification?: ProductSpecification;
    isNew: boolean;
}

export default function SpecificationEditForm({ specification, isNew }: SpecificationEditFormProps) {
    const navigate = useNavigate();
    const createMutation = useCreateSpecification();
    const updateMutation = useUpdateSpecification();

    const [name, setName] = useState(specification?.name || '');
    const [description, setDescription] = useState(specification?.description || '');
    const [productNumber, setProductNumber] = useState(specification?.productNumber || '');
    const [lifecycleStatus, setLifecycleStatus] = useState<LifecycleStatus>(specification?.lifecycleStatus || 'Draft');
    const [isBundle, setIsBundle] = useState(specification?.isBundle || false);
    const [characteristics, setCharacteristics] = useState<Record<string, ProductSpecCharacteristic>>(specification?.characteristics || {});

    const [startDate, setStartDate] = useState(specification?.validFor?.startDateTime?.split('T')[0] || '');
    const [endDate, setEndDate] = useState(specification?.validFor?.endDateTime?.split('T')[0] || '');

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        const payload = {
            name,
            description,
            productNumber,
            lifecycleStatus,
            isBundle,
            characteristics,
            validFor: {
                startDateTime: startDate ? `${startDate}T00:00:00Z` : undefined,
                endDateTime: endDate ? `${endDate}T23:59:59Z` : undefined,
            },
        };

        try {
            if (isNew) {
                const result = await createMutation.mutateAsync(payload as CreateProductSpecificationPayload);
                navigate(`/catalog/specifications/${result.id}`);
            } else if (specification) {
                await updateMutation.mutateAsync({
                    id: specification.id,
                    payload: payload as Partial<ProductSpecification>,
                });
                navigate(`/catalog/specifications/${specification.id}`);
            }
        } catch (err) {
            console.error('Failed to save specification:', err);
        }
    };

    const isPending = createMutation.isPending || updateMutation.isPending;

    return (
        <>
            <div className="page-header">
                <div className="page-header-content mb-3">
                    <h2>{isNew ? 'New Product Specification' : 'Edit Specification'}</h2>
                    <p className="lead text-muted mb-0">
                        {isNew ? 'Create a new technical blueprint' : `Editing ${specification?.name}`}
                    </p>
                </div>
                <button
                    type="submit"
                    form="spec-form"
                    className="btn btn-primary"
                    disabled={isPending}
                >
                    {isPending ? <Loader2 className="spin" size={18} /> : <Save size={18} />}
                    <span>{isNew ? 'Create Specification' : 'Save Changes'}</span>
                </button>
            </div>

            <form id="spec-form" className="form-container" onSubmit={handleSubmit}>
                <div className="card form-section">
                    <h3>Basic Information</h3>
                    <div className="form-grid">
                        <div className="form-group">
                            <label htmlFor="name">Name</label>
                            <input
                                id="name"
                                type="text"
                                value={name}
                                onChange={(e) => setName(e.target.value)}
                                required
                                placeholder="e.g., iPhone 15 Pro Max"
                            />
                        </div>
                        <div className="form-group">
                            <label htmlFor="productNumber">Product Number (SKU)</label>
                            <input
                                id="productNumber"
                                type="text"
                                value={productNumber}
                                onChange={(e) => setProductNumber(e.target.value)}
                                required
                                placeholder="e.g., AAPL-I15PM-256-BLK"
                            />
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
                        <div className="form-group" style={{ display: 'flex', alignItems: 'center', marginTop: '1.5rem' }}>
                            <label className="checkbox-label">
                                <input
                                    type="checkbox"
                                    checked={isBundle}
                                    onChange={(e) => setIsBundle(e.target.checked)}
                                />
                                <span>Is Bundle</span>
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
                            placeholder="Detailed description of the product specification..."
                        />
                    </div>
                </div>

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

                <CharacteristicEditor
                    characteristics={characteristics}
                    onChange={setCharacteristics}
                />
            </form>
        </>
    );
}
