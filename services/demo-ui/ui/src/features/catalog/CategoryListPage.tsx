import { useState, useMemo } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Plus, Edit, Trash2, ChevronRight, ChevronDown, FolderTree } from 'lucide-react';
import { EmptyState } from '../../design-system/components/common/EmptyState';
import { useCategories, useDeleteCategory } from './api';
import type { Category } from './types';
import './CategoryListPage.css';

interface CategoryNode extends Category {
    children: CategoryNode[];
    level: number;
}

export default function CategoryListPage() {
    const navigate = useNavigate();
    const { data: categories = [], isLoading, isError } = useCategories();
    const deleteMutation = useDeleteCategory();
    const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());

    // Build the tree structure
    const categoryTree = useMemo(() => {
        const buildTree = (parentId: string | undefined, level: number): CategoryNode[] => {
            return categories
                .filter(c => c.parentId === parentId || (!parentId && (!c.parentId || c.isRoot)))
                .map(c => ({
                    ...c,
                    level,
                    children: buildTree(c.id, level + 1)
                }));
        };
        return buildTree(undefined, 0);
    }, [categories]);

    const toggleExpand = (id: string) => {
        setExpandedIds(prev => {
            const next = new Set(prev);
            if (next.has(id)) next.delete(id);
            else next.add(id);
            return next;
        });
    };

    const handleDelete = async (id: string) => {
        if (confirm('Are you sure you want to delete this category?')) {
            await deleteMutation.mutateAsync(id);
        }
    };

    const renderNode = (node: CategoryNode) => {
        const hasChildren = node.children.length > 0;
        const isExpanded = expandedIds.has(node.id);

        return (
            <div key={node.id}>
                <div
                    className="category-item"
                    style={{ paddingLeft: `${node.level * 24 + 16}px` }}
                >
                    <div className="category-info">
                        {hasChildren ? (
                            <button
                                className="btn-icon-toggle"
                                onClick={() => toggleExpand(node.id)}
                                aria-label={isExpanded ? `Collapse ${node.name}` : `Expand ${node.name}`}
                            >
                                {isExpanded ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
                            </button>
                        ) : (
                            <span style={{ width: '24px' }}></span>
                        )}
                        <FolderTree size={18} className="text-muted" style={{ opacity: 0.7 }} />
                        <div className="category-name">
                            <span>{node.name}</span>
                            {node.description && <span className="category-desc">{node.description}</span>}
                        </div>
                        {node.isRoot && <span className="badge-root">ROOT</span>}
                    </div>

                    <div className="category-actions">
                        <Link to={`/catalog/categories/${node.id}/edit`} className="btn-outline btn-outline-primary" title="Edit">
                            <Edit size={14} />
                        </Link>
                        <button
                            className="btn-outline btn-outline-danger"
                            onClick={() => handleDelete(node.id)}
                            disabled={hasChildren}
                            title={hasChildren ? "Remove sub-categories first" : "Delete category"}
                            style={hasChildren ? { opacity: 0.5, cursor: 'not-allowed' } : {}}
                        >
                            <Trash2 size={14} />
                        </button>
                    </div>
                </div>
                {hasChildren && isExpanded && (
                    <div className="category-children">
                        {node.children.map(renderNode)}
                    </div>
                )}
            </div>
        );
    };

    if (isLoading) return <div className="p-5 text-center">Loading categories...</div>;
    if (isError) return <div className="alert alert-danger">Error loading categories.</div>;

    return (
        <div className="category-page">
            <div className="category-header">
                <div>
                    <h1 className="page-title">Categories</h1>
                    <p>Manage product catalog hierarchy</p>
                </div>
                <button className="btn btn-primary" onClick={() => navigate('/catalog/categories/new')}>
                    <Plus size={18} className="me-2" />
                    New Category
                </button>
            </div>

            <div className="category-tree-card">
                <div className="category-list">
                    {categoryTree.length === 0 ? (
                        <EmptyState
                            bare
                            icon={<FolderTree size={48} />}
                            title="No categories found."
                            action={
                                <button className="btn btn-primary" onClick={() => navigate('/catalog/categories/new')}>
                                    Create your first category
                                </button>
                            }
                        />
                    ) : (
                        categoryTree.map(renderNode)
                    )}
                </div>
            </div>
        </div>
    );
}
