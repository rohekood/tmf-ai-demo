import { useState, useMemo, useCallback } from 'react';
import { Link } from 'react-router-dom';
import {
    createColumnHelper,
    getCoreRowModel,
    getSortedRowModel,
    flexRender,
    type SortingState,
} from '@tanstack/react-table';
import { useReactTable } from '../../hooks/useReactTable';
import { Plus, Search, ChevronUp, ChevronDown, Eye, Edit, Trash2, Loader2, Package } from 'lucide-react';
import { useSpecifications, useDeleteSpecification } from './api';
import type { ProductSpecification } from './types';
import './SpecificationListPage.css';
import { IconButton } from '../../design-system/components/common/IconButton';
import { IconButtonArea } from '../../design-system/components/common/IconButtonArea';

const columnHelper = createColumnHelper<ProductSpecification>();

export default function SpecificationListPage() {
    const [searchTerm, setSearchTerm] = useState('');
    const [sorting, setSorting] = useState<SortingState>([]);

    const { data: specifications = [], isLoading, error } = useSpecifications();
    const deleteMutation = useDeleteSpecification();

    const handleDelete = useCallback((id: string, name: string) => {
        if (confirm(`Are you sure you want to delete specification "${name}"?`)) {
            deleteMutation.mutate(id);
        }
    }, [deleteMutation]);

    const filteredSpecs = useMemo(() => {
        if (!searchTerm) return specifications;
        const lowSearch = searchTerm.toLowerCase();
        return specifications.filter(s =>
            s.name.toLowerCase().includes(lowSearch) ||
            s.productNumber?.toLowerCase().includes(lowSearch)
        );
    }, [specifications, searchTerm]);

    const columns = useMemo(
        () => [
            columnHelper.accessor('name', {
                header: 'Name',
                cell: (info) => (
                    <div className="spec-info">
                        <span className="spec-name">{info.getValue()}</span>
                        {info.row.original.isBundle && (
                            <span className="bundle-badge">Bundle</span>
                        )}
                    </div>
                ),
            }),
            columnHelper.accessor('productNumber', {
                header: 'Product Number',
                cell: (info) => <span className="spec-number">{info.getValue()}</span>,
                size: 150,
            }),
            columnHelper.accessor('lifecycleStatus', {
                header: 'Status',
                cell: (info) => (
                    <span className={`status-badge ${info.getValue().toLowerCase()}`}>
                        {info.getValue()}
                    </span>
                ),
                size: 120,
            }),
            columnHelper.accessor((row) => Object.keys(row.characteristics || {}).length, {
                id: 'characteristics',
                header: 'Characteristics',
                cell: (info) => (
                    <span className="count-text">{info.getValue()} items</span>
                ),
                size: 130,
            }),
            columnHelper.display({
                id: 'actions',
                header: '',
                cell: (info) => (
                    <IconButtonArea alignment="end">
                        <IconButton
                            to={`/catalog/specifications/${info.row.original.id}`}
                            icon={<Eye size={16} />}
                            title="View"
                        />
                        <IconButton
                            to={`/catalog/specifications/${info.row.original.id}/edit`}
                            icon={<Edit size={16} />}
                            title="Edit"
                        />
                        <IconButton
                            variant="danger"
                            title="Delete"
                            onClick={() => handleDelete(info.row.original.id, info.row.original.name)}
                            disabled={deleteMutation.isPending}
                            icon={<Trash2 size={16} />}
                        />
                    </IconButtonArea>
                ),
                size: 120,
            }),
        ],
        [deleteMutation.isPending, handleDelete]
    );

    const table = useReactTable({
        data: filteredSpecs,
        columns,
        state: { sorting },
        onSortingChange: setSorting,
        getCoreRowModel: getCoreRowModel(),
        getSortedRowModel: getSortedRowModel(),
    });

    return (
        <div className="specification-list-page">
            <div className="page-header">
                <div className="page-header-content">
                    <h2>Product Specifications</h2>
                    <p className="page-description">Define the technical blueprints for your products</p>
                </div>
                <Link to="/catalog/specifications/new" className="btn btn-primary">
                    <Plus size={18} />
                    <span>New Specification</span>
                </Link>
            </div>

            <div className="search-bar card">
                <Search size={20} className="search-icon" />
                <input
                    type="text"
                    placeholder="Search by name or product number..."
                    className="search-input"
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                />
            </div>

            {error ? (
                <div className="card error-card" role="alert">
                    <p>Failed to load specifications: {error.message}</p>
                </div>
            ) : isLoading ? (
                <div className="card loading-card" role="status">
                    <Loader2 className="spin" size={24} />
                    <p>Loading specifications...</p>
                </div>
            ) : filteredSpecs.length === 0 ? (
                <div className="card empty-card" role="status">
                    <Package size={48} className="empty-icon" />
                    <p>{searchTerm ? 'No specifications match your search.' : 'No product specifications found.'}</p>
                    {!searchTerm && (
                        <Link to="/catalog/specifications/new" className="btn btn-primary">
                            Create your first specification
                        </Link>
                    )}
                </div>
            ) : (
                <div className="card table-card">
                    <table className="data-table">
                        <thead>
                            {table.getHeaderGroups().map((headerGroup) => (
                                <tr key={headerGroup.id}>
                                    {headerGroup.headers.map((header) => (
                                        <th
                                            key={header.id}
                                            onClick={header.column.getToggleSortingHandler()}
                                            className={header.column.getCanSort() ? 'sortable' : ''}
                                            style={{ width: header.getSize() }}
                                        >
                                            <div className="th-content">
                                                {flexRender(header.column.columnDef.header, header.getContext())}
                                                {header.column.getIsSorted() === 'asc' && <ChevronUp size={14} />}
                                                {header.column.getIsSorted() === 'desc' && <ChevronDown size={14} />}
                                            </div>
                                        </th>
                                    ))}
                                </tr>
                            ))}
                        </thead>
                        <tbody>
                            {table.getRowModel().rows.map((row) => (
                                <tr key={row.id}>
                                    {row.getVisibleCells().map((cell) => (
                                        <td key={cell.id}>
                                            {flexRender(cell.column.columnDef.cell, cell.getContext())}
                                        </td>
                                    ))}
                                </tr>
                            ))}
                        </tbody>
                    </table>
                    <div className="table-footer">
                        <span className="table-count">{filteredSpecs.length} specifications</span>
                    </div>
                </div>
            )}
        </div>
    );
}
