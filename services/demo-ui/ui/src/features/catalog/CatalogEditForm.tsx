import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Save, Loader2 } from 'lucide-react';
import { useCreateCatalog, useCatalogUpdate, useCategories, useUpdateCategory } from './api';
import type { Catalog, LifecycleStatus, CreateCatalogPayload } from './types';
import CategoryPicker from './components/CategoryPicker';
import '../parties/PartyFormPage.css';

interface CatalogEditFormProps {
    catalog?: Catalog;
    isNew: boolean;
}

export default function CatalogEditForm({ catalog, isNew }: CatalogEditFormProps) {
    const navigate = useNavigate();
    const createCatalogMutation = useCreateCatalog();
    const updateCatalogMutation = useCatalogUpdate();
    const [formError, setFormError] = useState<string | null>(null);

    const [name, setName] = useState(catalog?.name || '');
    const [description, setDescription] = useState(catalog?.description || '');
    const [lifecycleStatus, setLifecycleStatus] = useState<LifecycleStatus>(catalog?.lifecycleStatus || 'Draft');
    const [startDate, setStartDate] = useState(catalog?.validFor?.startDateTime?.split('T')[0] || '');
    const [endDate, setEndDate] = useState(catalog?.validFor?.endDateTime?.split('T')[0] || '');

    // Categories management
    const { data: allCategories = [] } = useCategories();
    // Categories linked to this catalog
    const catalogCategories = allCategories.filter(c => c.catalogId === catalog?.id);

    const updateCategoryMutation = useUpdateCategory();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        const payload = {
            name,
            description,
            lifecycleStatus,
            validFor: {
                startDateTime: startDate ? `${startDate}T00:00:00Z` : undefined,
                endDateTime: endDate ? `${endDate}T23:59:59Z` : undefined,
            },
        };

        setFormError(null);
        try {
            if (isNew) {
                const result = await createCatalogMutation.mutateAsync(payload as CreateCatalogPayload);
                navigate(`/catalog/catalogs/${result.id}`);
            } else if (catalog) {
                await updateCatalogMutation.mutateAsync({
                    id: catalog.id,
                    payload: payload as Partial<Catalog>,
                });
                navigate(`/catalog/catalogs/${catalog.id}`);
            }
        } catch (err: unknown) {
            const message = err instanceof Error ? err.message : 'Failed to save catalog.';
            setFormError(message);
        }
    };

    const isPending = createCatalogMutation.isPending || updateCatalogMutation.isPending;

    return (
        <>
            <div className="page-header">
                <div className="page-header-content mb-3">
                    <h2>{isNew ? 'New Product Catalog' : 'Edit Catalog'}</h2>
                    <p className="lead text-muted mb-0">
                        {isNew ? 'Define a new collection' : `Editing ${catalog?.name}`}
                    </p>
                </div>
                <button
                    type="submit"
                    form="catalog-form"
                    className="btn btn-primary"
                    disabled={isPending}
                >
                    {isPending ? <Loader2 className="spin" size={18} /> : <Save size={18} />}
                    <span>{isNew ? 'Create Catalog' : 'Save Changes'}</span>
                </button>
            </div>

            {formError && (
                <div className="card error-card" role="alert" style={{ marginBottom: '1rem' }}>
                    <p>{formError}</p>
                </div>
            )}

            <form id="catalog-form" className="form-container" onSubmit={handleSubmit}>
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
                                placeholder="e.g., Enterprise Product Catalog"
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
                    </div>
                    <div className="form-group mt-3">
                        <label htmlFor="description">Description</label>
                        <textarea
                            id="description"
                            value={description}
                            onChange={(e) => setDescription(e.target.value)}
                            rows={3}
                            placeholder="Describe the purpose of this catalog..."
                        />
                    </div>
                </div>

                <div className="card form-section">
                    <h3>Validity Period</h3>
                    <div className="form-grid">
                        <div className="form-group">
                            <label htmlFor="startDate">Start Date</label>
                            <input
                                id="startDate"
                                type="date"
                                value={startDate}
                                onChange={(e) => setStartDate(e.target.value)}
                            />
                        </div>
                        <div className="form-group">
                            <label htmlFor="endDate">End Date</label>
                            <input
                                id="endDate"
                                type="date"
                                value={endDate}
                                onChange={(e) => setEndDate(e.target.value)}
                            />
                        </div>
                    </div>
                </div>

                {!isNew && (
                    <div className="card form-section">
                        <CategoryPicker
                            selectedIds={catalogCategories.map(c => c.id)}
                            onChange={async (newIds) => {
                                if (!catalog) return;
                                const oldIds = catalogCategories.map(c => c.id);
                                const added = newIds.filter(id => !oldIds.includes(id));
                                const removed = oldIds.filter(id => !newIds.includes(id));

                                try {
                                    for (const id of added) {
                                        await updateCategoryMutation.mutateAsync({
                                            id,
                                            payload: { catalogId: catalog.id }
                                        });
                                    }
                                    for (const id of removed) {
                                        await updateCategoryMutation.mutateAsync({
                                            id,
                                            payload: { catalogId: null }
                                        });
                                    }
                                } catch (error) {
                                    console.error("Failed to update category links", error);
                                }
                            }}
                            variant="list"
                            categories={allCategories}
                        />
                    </div>
                )}
            </form>
        </>
    );
}
