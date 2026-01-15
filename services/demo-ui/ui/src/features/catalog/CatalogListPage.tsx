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
import { Plus, Search, ChevronUp, ChevronDown, Eye, Edit, Trash2, Loader2, Book } from 'lucide-react';
import { useCatalogs, useCatalogDelete } from './api';
import type { Catalog } from './types';
import './SpecificationListPage.css'; // Reuse common catalog styles
import { IconButton } from '../../design-system/components/common/IconButton';
import { IconButtonArea } from '../../design-system/components/common/IconButtonArea';

const columnHelper = createColumnHelper<Catalog>();

export default function CatalogListPage() {
    const [searchTerm, setSearchTerm] = useState('');
    const [sorting, setSorting] = useState<SortingState>([]);

    const { data: catalogs = [], isLoading, error } = useCatalogs();
    const deleteMutation = useCatalogDelete();

    const handleDelete = useCallback((id: string, name: string) => {
        if (confirm(`Are you sure you want to delete catalog "${name}"?`)) {
            deleteMutation.mutate(id);
        }
    }, [deleteMutation]);

    const filteredCatalogs = useMemo(() => {
        if (!searchTerm) return catalogs;
        const lowSearch = searchTerm.toLowerCase();
        return catalogs.filter(c =>
            c.name.toLowerCase().includes(lowSearch) ||
            c.description?.toLowerCase().includes(lowSearch)
        );
    }, [catalogs, searchTerm]);

    const columns = useMemo(
        () => [
            columnHelper.accessor('name', {
                header: 'Name',
                cell: (info) => <span className="spec-name">{info.getValue()}</span>,
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
            columnHelper.accessor('lastUpdate', {
                header: 'Last Updated',
                cell: (info) => <span className="count-text">{new Date(info.getValue()).toLocaleDateString()}</span>,
                size: 150,
            }),
            columnHelper.display({
                id: 'actions',
                header: '',
                cell: (info) => (
                    <IconButtonArea alignment="end">
                        <IconButton
                            to={`/catalog/catalogs/${info.row.original.id}`}
                            icon={<Eye size={16} />}
                            title="View"
                        />
                        <IconButton
                            to={`/catalog/catalogs/${info.row.original.id}/edit`}
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
        data: filteredCatalogs,
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
                    <h2>Product Catalogs</h2>
                    <p className="page-description">Organizational collections of product offerings</p>
                </div>
                <Link to="/catalog/catalogs/new" className="btn btn-primary">
                    <Plus size={18} />
                    <span>New Catalog</span>
                </Link>
            </div>

            <div className="search-bar card">
                <Search size={20} className="search-icon" />
                <input
                    type="text"
                    placeholder="Search by name or description..."
                    className="search-input"
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                />
            </div>

            {error ? (
                <div className="card error-card" role="alert">
                    <p>Failed to load catalogs: {error.message}</p>
                </div>
            ) : isLoading ? (
                <div className="card loading-card" role="status">
                    <Loader2 className="spin" size={24} />
                    <p>Loading catalogs...</p>
                </div>
            ) : filteredCatalogs.length === 0 ? (
                <div className="card empty-card" role="status">
                    <Book size={48} className="empty-icon" />
                    <p>{searchTerm ? 'No catalogs match your search.' : 'No product catalogs found.'}</p>
                    {!searchTerm && (
                        <Link to="/catalog/catalogs/new" className="btn btn-primary">
                            Create your first catalog
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
                        <span className="table-count">{filteredCatalogs.length} catalogs</span>
                    </div>
                </div>
            )}
        </div>
    );
}
