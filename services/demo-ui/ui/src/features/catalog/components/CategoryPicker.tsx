import { useState } from 'react';
import { Tag, X, Plus, Search, Link as LinkIcon, Unlink } from 'lucide-react';
import { useCategories } from '../api';
import type { Category } from '../types';
import { IconButton } from '../../../design-system/components/common/IconButton';
import './CategoryPicker.css';

interface CategoryPickerProps {
    selectedIds: string[];
    onChange: (ids: string[]) => void;
    catalogId?: string;
    variant?: 'tags' | 'list';
    categories?: Category[];
}

export default function CategoryPicker({
    selectedIds,
    onChange,
    catalogId,
    variant = 'tags',
    categories: providedCategories
}: CategoryPickerProps) {
    const { data: fetchedCategories = [], isLoading } = useCategories();
    // Use provided categories if available, otherwise use fetched ones
    const allCategories = providedCategories || fetchedCategories;

    const [isSelecting, setIsSelecting] = useState(false);
    const [searchTerm, setSearchTerm] = useState('');

    const filteredCategories = catalogId
        ? allCategories.filter(c => c.catalogId === catalogId)
        : allCategories;

    const selectedCategories = allCategories.filter(c => selectedIds.includes(c.id));
    const availableCategories = filteredCategories
        .filter(c => !selectedIds.includes(c.id))
        .filter(c => c.name.toLowerCase().includes(searchTerm.toLowerCase()));

    const handleSelect = (id: string) => {
        onChange([...selectedIds, id]);
        setIsSelecting(false);
        setSearchTerm('');
    };

    const handleRemove = (id: string) => {
        onChange(selectedIds.filter(i => i !== id));
    };

    if (variant === 'list') {
        return (
            <div className="category-picker-list">
                <div className="section-header d-flex justify-content-between align-items-center mb-3">
                    <h4 className="m-0">Categories</h4>
                    {!isSelecting ? (
                        <button
                            type="button"
                            className="btn btn-secondary btn-sm"
                            onClick={() => setIsSelecting(true)}
                            disabled={isLoading}
                        >
                            <LinkIcon size={16} className="me-2" />
                            <span>Link Category</span>
                        </button>
                    ) : (
                        <button type="button" className="btn btn-secondary btn-sm" onClick={() => setIsSelecting(false)}>
                            <X size={16} className="me-2" />
                            <span>Cancel</span>
                        </button>
                    )}
                </div>

                {isSelecting && (
                    <div className="mb-3 p-3 border rounded bg-light">
                        <div className="input-group mb-2">
                            <span className="input-group-text bg-white">
                                <Search size={16} />
                            </span>
                            <input
                                type="text"
                                className="form-control"
                                placeholder="Search categories..."
                                value={searchTerm}
                                onChange={e => setSearchTerm(e.target.value)}
                                autoFocus
                            />
                        </div>
                        <div className="list-group" style={{ maxHeight: '200px', overflowY: 'auto' }}>
                            {availableCategories.length === 0 ? (
                                <div className="p-2 text-muted text-center">No available categories found</div>
                            ) : (
                                availableCategories.map(cat => (
                                    <button
                                        key={cat.id}
                                        type="button"
                                        className="list-group-item list-group-item-action d-flex justify-content-between align-items-center"
                                        onClick={() => handleSelect(cat.id)}
                                    >
                                        <span>{cat.name}</span>
                                        {cat.catalogId && <small className="text-muted">Currently in another catalog</small>}
                                    </button>
                                ))
                            )}
                        </div>
                    </div>
                )}

                <div className="category-list">
                    {selectedCategories.length === 0 ? (
                        <div className="category-list-empty">
                            <p className="text-muted fst-italic m-0">No categories linked.</p>
                        </div>
                    ) : (
                        selectedCategories.map(cat => (
                            <div key={cat.id} className="category-list-item">
                                <div className="category-name">
                                    <strong>{cat.name}</strong>
                                </div>
                                <div className="category-status">
                                    <span className={`status-badge status-${cat.lifecycleStatus.toLowerCase()}`}>
                                        {cat.lifecycleStatus}
                                    </span>
                                </div>
                                <div className="category-actions">
                                    <IconButton
                                        icon={<Unlink size={16} />}
                                        variant="danger"
                                        size="sm"
                                        onClick={() => handleRemove(cat.id)}
                                        title="Unlink Category"
                                    />
                                </div>
                            </div>
                        ))
                    )}
                </div>
            </div>
        );
    }

    return (
        <div className="category-picker">
            <div className="selected-categories">
                {selectedCategories.map(cat => (
                    <span key={cat.id} className="category-tag">
                        <Tag size={14} />
                        {cat.name}
                        <button type="button" onClick={() => handleRemove(cat.id)} aria-label={`Remove ${cat.name}`}>
                            <X size={14} />
                        </button>
                    </span>
                ))}
                {!isSelecting && !isLoading && (
                    <button type="button" className="btn-secondary" onClick={() => setIsSelecting(true)}>
                        <Plus size={14} />
                        <span>Add to Category</span>
                    </button>
                )}
            </div>

            {isSelecting && (
                <div className="picker-dropdown">
                    <div className="picker-header">
                        <h4>Select Category</h4>
                        <button type="button" className="btn-icon btn-sm" onClick={() => setIsSelecting(false)}>
                            <X size={16} />
                        </button>
                    </div>
                    <div className="picker-search p-2 border-bottom">
                        <input
                            type="text"
                            placeholder="Search categories..."
                            className="form-control"
                            value={searchTerm}
                            onChange={(e) => setSearchTerm(e.target.value)}
                            autoFocus
                        />
                    </div>
                    {availableCategories.length === 0 ? (
                        <p className="p-3 text-muted">No additional categories available.</p>
                    ) : (
                        <div className="picker-list">
                            {availableCategories.map(cat => (
                                <button
                                    key={cat.id}
                                    type="button"
                                    className="picker-item"
                                    onClick={() => handleSelect(cat.id)}
                                >
                                    <strong>{cat.name}</strong>
                                    {cat.catalogId && <span className="text-muted ml-2" style={{ fontSize: '0.75rem' }}>in {cat.catalogId}</span>}
                                </button>
                            ))}
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}
