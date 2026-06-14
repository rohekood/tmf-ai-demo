import { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Save, Loader2, ArrowLeft } from 'lucide-react';
import { useCategory, useCreateCategory, useUpdateCategory, useCategories } from './api';
import type { Category, CreateCategoryPayload, LifecycleStatus } from './types';
import '../parties/PartyFormPage.css'; // Reusing styles
import './CategoryEditPage.css';

interface CategoryFormProps {
    initialData?: Category;
    onSubmit: (data: CreateCategoryPayload | Partial<Category>) => Promise<void>;
    isPending: boolean;
    availableParents: Category[];
}

function CategoryForm({ initialData, onSubmit, isPending, availableParents }: CategoryFormProps) {
    const [name, setName] = useState(initialData?.name || '');
    const [description, setDescription] = useState(initialData?.description || '');
    const [lifecycleStatus, setLifecycleStatus] = useState<LifecycleStatus>(initialData?.lifecycleStatus || 'Draft');
    const [isRoot, setIsRoot] = useState(initialData?.isRoot || false);
    const [parentId, setParentId] = useState<string>(initialData?.parentId || '');
    const [startDate, setStartDate] = useState(initialData?.validFor?.startDateTime?.split('T')[0] || '');
    const [endDate, setEndDate] = useState(initialData?.validFor?.endDateTime?.split('T')[0] || '');

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();

        const payload = {
            name,
            description,
            lifecycleStatus,
            isRoot,
            parentId: isRoot ? undefined : parentId,
            validFor: {
                startDateTime: startDate ? `${startDate}T00:00:00Z` : undefined,
                endDateTime: endDate ? `${endDate}T23:59:59Z` : undefined,
            },
        };

        onSubmit(payload);
    };

    return (
        <form id="category-form" className="form-container" onSubmit={handleSubmit}>
            <fieldset disabled={isPending} style={{ border: 'none', padding: 0, margin: 0, display: 'contents' }}>
                <div className="card form-section">
                    <h3>Basic Information</h3>
                    <div className="form-grid">
                        <div className="form-group">
                            <label htmlFor="name">Category Name</label>
                            <input
                                id="name"
                                type="text"
                                value={name}
                                onChange={(e) => setName(e.target.value)}
                                required
                                placeholder="e.g., Electronics"
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
                            placeholder="Description of this category..."
                        />
                    </div>
                </div>

                <div className="card form-section">
                    <h3>Hierarchy</h3>
                    <div className="form-group mb-3">
                        <label className="checkbox-label">
                            <input
                                type="checkbox"
                                checked={isRoot}
                                onChange={(e) => {
                                    setIsRoot(e.target.checked);
                                    if (e.target.checked) setParentId('');
                                }}
                            />
                            <span>Is Root Category?</span>
                        </label>
                        <small className="form-text text-muted d-block ms-4">
                            Root categories appear at the top level of the catalog.
                        </small>
                    </div>

                    {!isRoot && (
                        <div className="form-group">
                            <label htmlFor="parent">Parent Category</label>
                            <select
                                id="parent"
                                value={parentId}
                                onChange={(e) => setParentId(e.target.value)}
                                required={!isRoot}
                            >
                                <option value="">Select a parent...</option>
                                {availableParents.map(c => (
                                    <option key={c.id} value={c.id}>{c.name}</option>
                                ))}
                            </select>
                        </div>
                    )}
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
            </fieldset>
        </form>
    );
}

export default function CategoryEditPage() {
    const { id } = useParams<{ id: string }>();
    const isNew = !id;
    const navigate = useNavigate();
    const [formError, setFormError] = useState<string | null>(null);

    const { data: category, isLoading: isCategoryLoading } = useCategory(id);
    const { data: allCategories = [] } = useCategories();

    const createMutation = useCreateCategory();
    const updateMutation = useUpdateCategory();

    const handleSubmit = async (payload: CreateCategoryPayload | Partial<Category>) => {
        setFormError(null);
        try {
            if (isNew) {
                await createMutation.mutateAsync(payload as CreateCategoryPayload);
            } else if (id) {
                await updateMutation.mutateAsync({
                    id,
                    payload: payload as Partial<Category>,
                });
            }
            navigate('/catalog/categories');
        } catch (err: unknown) {
            const message = err instanceof Error ? err.message : 'Failed to save category.';
            setFormError(message);
        }
    };

    const isPending = createMutation.isPending || updateMutation.isPending;
    const isLoading = !isNew && isCategoryLoading;

    // Filter potential parents to prevent self-selection
    const availableParents = allCategories.filter(c => c.id !== id);

    if (isLoading) return <div className="p-5 text-center"><Loader2 className="spin" size={32} /></div>;

    return (
        <div className="container-fluid py-4">
            <div className="mb-4">
                <button className="btn btn-link mb-2" style={{ display: 'inline-flex', alignItems: 'center' }} onClick={() => navigate('/catalog/categories')}>
                    <ArrowLeft size={16} className="me-1" />
                    Back to Categories
                </button>
            </div>

            <div className="page-header">
                <div className="page-header-content mb-3">
                    <h2>{isNew ? 'New Category' : 'Edit Category'}</h2>
                    <p className="lead text-muted mb-0">
                        {isNew ? 'Create a product category' : `Editing ${category?.name}`}
                    </p>
                </div>
                <button
                    type="submit"
                    form="category-form"
                    className="btn btn-primary"
                    disabled={isPending}
                >
                    {isPending ? <Loader2 className="spin" size={18} /> : <Save size={18} />}
                    <span>{isNew ? 'Create Category' : 'Save Changes'}</span>
                </button>
            </div>

            {formError && (
                <div className="alert alert-danger" role="alert" style={{ marginBottom: '1rem' }}>
                    {formError}
                </div>
            )}

            <CategoryForm
                initialData={category}
                onSubmit={handleSubmit}
                isPending={isPending}
                availableParents={availableParents}
            />
        </div>
    );
}
